package dockerapiagentuc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"sync"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerproxy"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// socketMode lets an app running as any user reach its socket. Which apps
	// reach it is decided by who mounts the volume it lives in.
	socketMode = 0o666
	// readHeaderTimeout bounds a client that connects and says nothing. Nothing
	// else is bounded: logs and attach are streams.
	readHeaderTimeout = 10 * time.Second
)

// socketHost keeps one proxy socket per app with access, on this node.
type socketHost struct {
	mu      sync.Mutex
	sockets map[string]*appSocket

	logger   logging.Logger
	upstream http.RoundTripper
	// socketDir is where an app's socket goes on this node: its socket volume,
	// reached through the host's filesystem.
	socketDir func(ctx context.Context, policy *dockerproxy.Policy) (string, error)
}

type appSocket struct {
	proxy  *dockerproxy.Proxy
	server *http.Server
	path   string
}

func newSocketHost(logger logging.Logger, dockerManager docker.Manager, upstream http.RoundTripper) *socketHost {
	return &socketHost{
		sockets:  map[string]*appSocket{},
		logger:   logger,
		upstream: upstream,
		socketDir: func(ctx context.Context, policy *dockerproxy.Policy) (string, error) {
			return volumeSocketDir(ctx, dockerManager, policy)
		},
	}
}

// newUpstream reaches the daemon the way the docker CLI would, from the
// environment: on a node, its socket.
func newUpstream() (http.RoundTripper, error) {
	daemon, err := client.New(client.FromEnv)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	dial := daemon.Dialer()
	return &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return dial(ctx)
	}}, nil
}

// reconcile serves exactly the apps of policies. A new app gets a socket, an
// app whose policy changed has the new one from its next request, and an app no
// longer listed loses its socket. One app that cannot be served does not stop
// the others.
func (h *socketHost) reconcile(ctx context.Context, policies []*dockerproxy.Policy) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	wanted := make(map[string]bool, len(policies))
	var errs []error
	for _, policy := range policies {
		wanted[policy.AppID] = true
		if socket, found := h.sockets[policy.AppID]; found {
			socket.proxy.SetPolicy(policy)
			continue
		}
		if err := h.open(ctx, policy); err != nil {
			errs = append(errs, err)
		}
	}
	for appID, socket := range h.sockets {
		if !wanted[appID] {
			socket.close()
			delete(h.sockets, appID)
		}
	}
	return errors.Join(errs...)
}

func (h *socketHost) open(ctx context.Context, policy *dockerproxy.Policy) error {
	dir, err := h.socketDir(ctx, policy)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, dockerproxy.SocketFile)
	// A socket file outlives the agent that made it, and nothing can listen
	// where it is.
	if err = os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return hperrors.Wrap(err)
	}
	listener, err := (&net.ListenConfig{}).Listen(ctx, "unix", path)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if err = os.Chmod(path, socketMode); err != nil {
		_ = listener.Close()
		return hperrors.Wrap(err)
	}
	proxy := dockerproxy.New(policy, dockerproxy.Options{Upstream: h.upstream, OnDecision: h.logDecision})
	server := &http.Server{Handler: proxy, ReadHeaderTimeout: readHeaderTimeout}
	safego.GoWithLogger(h.logger, "dockerAPI.serve", func() {
		_ = server.Serve(listener)
	})
	h.sockets[policy.AppID] = &appSocket{proxy: proxy, server: server, path: path}
	return nil
}

func (s *appSocket) close() {
	_ = s.server.Close()
	_ = os.Remove(s.path)
}

// closeApp stops serving one app at once, rather than at the next reconcile.
func (h *socketHost) closeApp(appID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if socket, found := h.sockets[appID]; found {
		socket.close()
		delete(h.sockets, appID)
	}
}

// closeAll stops serving every app.
func (h *socketHost) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for appID, socket := range h.sockets {
		socket.close()
		delete(h.sockets, appID)
	}
}

// served are the apps this node serves, sorted.
func (h *socketHost) served() []string {
	h.mu.Lock()
	defer h.mu.Unlock()
	apps := make([]string, 0, len(h.sockets))
	for appID := range h.sockets {
		apps = append(apps, appID)
	}
	sort.Strings(apps)
	return apps
}

func (h *socketHost) logDecision(d dockerproxy.Decision) {
	if d.Allowed {
		h.logger.Debugf("docker api: app %s: %s %s: %s", d.AppID, d.Method, d.Path, d.Reason)
		return
	}
	h.logger.Warnf("docker api: app %s refused %s %s: %s", d.AppID, d.Method, d.Path, d.Reason)
}

// volumeSocketDir is where an app's socket lives on this node: its socket
// volume. Creating a volume that exists returns it, so this also finds one that
// Swarm made first, when it started the app's task.
func volumeSocketDir(ctx context.Context, dockerManager docker.Manager, policy *dockerproxy.Policy) (string, error) {
	resp, err := dockerManager.VolumeCreate(ctx, func(opts *client.VolumeCreateOptions) {
		opts.Name = policy.SocketVolume
		opts.Labels = map[string]string{dockerapiservice.SocketVolumeLabel: policy.AppID}
	})
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	// Creating a volume that exists returns it with the labels it already has.
	// One under this name that is not this app's was made by something else, and
	// serving this app's socket into it would put the socket where that
	// something else can read it.
	if owner := resp.Volume.Labels[dockerapiservice.SocketVolumeLabel]; owner != policy.AppID {
		return "", hperrors.Wrap(hperrors.ErrInfraInvalidArgument).
			WithMsgLog("socket volume %s is labeled for %q, not for app %s",
				policy.SocketVolume, owner, policy.AppID)
	}
	mountpoint := resp.Volume.Mountpoint
	// The agent's container reaches the host's filesystem under a prefix; an
	// agent running on the host itself reaches it as it is.
	candidates := []string{filepath.Join(volumeservice.HostPathPrefix, mountpoint), mountpoint}
	if i := slices.IndexFunc(candidates, isDir); i >= 0 {
		return candidates[i], nil
	}
	return "", hperrors.Wrap(hperrors.ErrInfraNotFound).
		WithMsgLog("socket volume %s is not reachable at %s", policy.SocketVolume, mountpoint)
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
