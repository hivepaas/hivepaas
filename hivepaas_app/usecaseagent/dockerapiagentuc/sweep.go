package dockerapiagentuc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// exitedChildAge is how long an exited child is kept for the app to read.
	exitedChildAge = 24 * time.Hour
	// idleNetworkAge is how long a network of the app's may stand empty: a job
	// makes its network before its first container.
	idleNetworkAge = time.Hour
)

// finishedStates are the states of a child that is not coming back by itself.
var finishedStates = []container.ContainerState{container.StateCreated, container.StateExited, container.StateDead}

// Removed counts what a sweep took away.
type Removed struct {
	Containers int
	Networks   int
	Volumes    int
}

// sweep removes, on this node, what apps' children left: everything of an app
// that is gone and, when ageOut is set, an exited child older than a day and an
// empty network older than an hour of an app that stays. Volumes of an app that
// stays are kept, since they are its caches.
type sweep struct {
	docker docker.Manager
	now    time.Time
	// gone reports an app whose every object goes.
	gone   func(appID string) bool
	ageOut bool
	// only, when set, limits the sweep to one app's objects.
	only string
}

func (s *sweep) run(ctx context.Context) (Removed, error) {
	var removed Removed
	var errs []error
	var err error
	removed.Containers, err = s.containers(ctx)
	errs = append(errs, err)
	removed.Networks, err = s.networks(ctx)
	errs = append(errs, err)
	volumes, err := s.volumes(ctx)
	errs = append(errs, err)
	sockets, err := s.socketVolumes(ctx)
	errs = append(errs, err)
	removed.Volumes = volumes + sockets
	return removed, errors.Join(errs...)
}

// ownerFilter selects the objects the proxy created, of one app when the sweep
// is limited to it.
func (s *sweep) ownerFilter() string {
	if s.only != "" {
		return dockerproxy.OwnerLabel + "=" + s.only
	}
	return dockerproxy.OwnerLabel
}

func (s *sweep) containers(ctx context.Context) (int, error) {
	resp, err := s.docker.ContainerList(ctx, func(opts *client.ContainerListOptions) {
		opts.All = true
		docker.FilterAdd(&opts.Filters, "label", s.ownerFilter())
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	count := 0
	var errs []error
	for _, c := range resp.Items {
		stale := s.gone(c.Labels[dockerproxy.OwnerLabel]) ||
			(s.ageOut && slices.Contains(finishedStates, c.State) &&
				s.now.Sub(time.Unix(c.Created, 0)) > exitedChildAge)
		if !stale {
			continue
		}
		_, err = s.docker.ContainerRemove(ctx, c.ID, func(opts *client.ContainerRemoveOptions) {
			opts.Force = true
			opts.RemoveVolumes = true
		})
		if err != nil {
			errs = append(errs, hperrors.Wrap(err))
			continue
		}
		count++
	}
	return count, errors.Join(errs...)
}

func (s *sweep) networks(ctx context.Context) (int, error) {
	resp, err := s.docker.NetworkList(ctx, func(opts *client.NetworkListOptions) {
		docker.FilterAdd(&opts.Filters, "label", s.ownerFilter())
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	count := 0
	var errs []error
	for _, n := range resp.Items {
		if !s.gone(n.Labels[dockerproxy.OwnerLabel]) {
			idle, idleErr := s.idleNetwork(ctx, n.ID, n.Created)
			if idleErr != nil {
				errs = append(errs, idleErr)
			}
			if !idle {
				continue
			}
		}
		if _, err = s.docker.NetworkRemove(ctx, n.ID); err != nil {
			errs = append(errs, hperrors.Wrap(err))
			continue
		}
		count++
	}
	return count, errors.Join(errs...)
}

// idleNetwork reports a network of an app that stays, old enough and with
// nothing attached, when the sweep ages things out.
func (s *sweep) idleNetwork(ctx context.Context, id string, created time.Time) (bool, error) {
	if !s.ageOut || s.now.Sub(created) <= idleNetworkAge {
		return false, nil
	}
	inspect, err := s.docker.NetworkInspect(ctx, id)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(inspect.Network.Containers) == 0, nil
}

func (s *sweep) volumes(ctx context.Context) (int, error) {
	resp, err := s.docker.VolumeList(ctx, func(opts *client.VolumeListOptions) {
		docker.FilterAdd(&opts.Filters, "label", s.ownerFilter())
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	var names []string
	for _, v := range resp.Items {
		if s.gone(v.Labels[dockerproxy.OwnerLabel]) {
			names = append(names, v.Name)
		}
	}
	return s.removeVolumes(ctx, names)
}

// socketVolumes removes the socket volumes of apps that are gone. They are
// found by name: one that Swarm created when it started the app's task carries
// no label.
func (s *sweep) socketVolumes(ctx context.Context) (int, error) {
	prefix := dockerapiservice.SocketVolumePrefix
	if s.only != "" {
		prefix = dockerapiservice.SocketVolumeName(s.only)
	}
	resp, err := s.docker.VolumeList(ctx, func(opts *client.VolumeListOptions) {
		docker.FilterAdd(&opts.Filters, "name", prefix)
	})
	if err != nil {
		return 0, hperrors.Wrap(err)
	}
	var names []string
	for _, v := range resp.Items {
		appID, isSocket := strings.CutPrefix(v.Name, dockerapiservice.SocketVolumePrefix)
		if isSocket && s.gone(appID) {
			names = append(names, v.Name)
		}
	}
	return s.removeVolumes(ctx, names)
}

// removeVolumes removes what it can. A volume still mounted - by the task of an
// app being deleted - stays until a later sweep.
func (s *sweep) removeVolumes(ctx context.Context, names []string) (int, error) {
	count := 0
	var errs []error
	for _, name := range names {
		if _, err := s.docker.VolumeRemove(ctx, name, false); err != nil {
			errs = append(errs, hperrors.Wrap(err))
			continue
		}
		count++
	}
	return count, errors.Join(errs...)
}
