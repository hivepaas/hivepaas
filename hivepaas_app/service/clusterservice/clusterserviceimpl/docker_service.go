package clusterserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	itemRemovalRetryDelay = 3 * time.Second
	itemRemovalRetryIncr  = 3 * time.Second
)

func (s *service) ServiceInspect(
	ctx context.Context,
	serviceID string,
	caching bool,
) (*swarm.Service, error) {
	if serviceID == "" {
		return nil, nil
	}

	// TODO: handle caching flag

	resp, err := s.dockerManager.ServiceInspect(ctx, serviceID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &resp.Service, nil
}

func (s *service) ServiceUpdate(
	ctx context.Context,
	serviceID string,
	version *swarm.Version,
	service *swarm.ServiceSpec,
	options ...docker.ServiceUpdateOption,
) (*client.ServiceUpdateResult, error) {
	if serviceID == "" {
		return nil, nil
	}
	resp, err := s.dockerManager.ServiceUpdate(ctx, serviceID, version, service, options...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return resp, nil
}

func (s *service) ServiceRemove(
	ctx context.Context,
	serviceID string,
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if serviceID == "" {
		return nil
	}
	fn := func() error {
		_, err := s.dockerManager.ServiceRemove(ctx, serviceID)
		if err != nil {
			if errors.Is(err, hperrors.ErrNotFound) {
				return nil
			}
			return hperrors.Wrap(err)
		}
		return nil
	}
	if retryMax > 0 {
		if retryDelay <= 0 {
			retryDelay = itemRemovalRetryDelay
		}
		err = gofn.ExecRetryCtx(ctx, fn, retryMax, retryDelay, gofn.ExecRetryDelayIncr(itemRemovalRetryIncr))
	} else {
		err = fn()
	}
	if err != nil {
		// TODO: create a cleanup task
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) ServicesRemove(
	ctx context.Context,
	serviceIDs []string,
	retryMax int,
	retryDelay time.Duration,
) (err error) {
	if len(serviceIDs) == 0 {
		return nil
	}
	if len(serviceIDs) == 1 {
		return s.ServiceRemove(ctx, serviceIDs[0], retryMax, retryDelay)
	}
	errMap := gofn.ExecTaskFuncEx(ctx, 10, false, //nolint:mnd
		func(ctx context.Context, itemID string) error {
			return s.ServiceRemove(ctx, itemID, retryMax, retryDelay)
		}, serviceIDs...)
	for _, e := range errMap {
		err = errors.Join(err, e)
	}
	return err
}

// VerifyPortsAvailable refuses a port that another service already publishes.
//
// Docker refuses a second service on the same ingress port itself, but only once
// the service is being created - by then an app has been provisioned, and what
// comes back is a message about swarm rather than about the port somebody asked
// for. Asking first is what lets a template creating three apps stop before it
// creates any of them, the way a taken domain already does.
//
// Host-mode ports are checked as well, even though docker would allow two of
// them on a cluster with more than one node: the tasks would then fight over the
// port on whichever node they land on, and a HivePaaS installation is usually
// one node, where they always would.
func (s *service) VerifyPortsAvailable(
	ctx context.Context,
	ports []clusterservice.PortRef,
	ignoreServiceIDs []string,
) error {
	if len(ports) == 0 {
		return nil
	}
	services, err := s.dockerManager.ServiceList(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	taken := make(map[clusterservice.PortRef]string, len(services.Items))
	for i := range services.Items {
		svc := &services.Items[i]
		if gofn.Contain(ignoreServiceIDs, svc.ID) || svc.Spec.EndpointSpec == nil {
			continue
		}
		for _, port := range svc.Spec.EndpointSpec.Ports {
			if port.PublishedPort == 0 {
				continue
			}
			ref := normalizePortRef(clusterservice.PortRef{
				Published: port.PublishedPort, Protocol: port.Protocol,
			})
			if _, found := taken[ref]; !found {
				taken[ref] = gofn.Coalesce(svc.Spec.Name, svc.ID)
			}
		}
	}
	for _, port := range ports {
		if by, found := taken[normalizePortRef(port)]; found {
			return hperrors.Wrap(hperrors.ErrPortInUse).
				WithParam("Port", port.Published).
				WithParam("Protocol", string(normalizePortRef(port).Protocol)).
				WithParam("PublishedBy", by)
		}
	}
	return nil
}

// normalizePortRef fills in the protocol docker assumes when a port is asked for
// without one, so that a tcp port and a port with no protocol are one address
// rather than two.
func normalizePortRef(port clusterservice.PortRef) clusterservice.PortRef {
	if port.Protocol == "" {
		port.Protocol = network.TCP
	}
	return port
}
