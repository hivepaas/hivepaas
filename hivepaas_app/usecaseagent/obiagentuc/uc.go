// Package obiagentuc runs OBI on the agent's node for the apps that ask for
// their routes and calls, and writes what it measures to the agent's stdout,
// where the log collector takes it with the agent's other rows. See
// docs/superpowers/specs/2026-10-03-obi-calls-and-routes-design.md.
package obiagentuc

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// reconcileInterval is how often the settings are read and OBI made to
	// match them: a switch turned takes this long at most.
	reconcileInterval = 30 * time.Second
	// scrapeInterval is how often OBI's metrics are read: a row per series
	// per interval, as the agent's resource rows.
	scrapeInterval = 15 * time.Second
	// statusInterval is how often the node says what it can run and runs.
	statusInterval = time.Minute
	// preflightInterval is how often the node is checked again: a kernel
	// does not change while it runs, its free memory does.
	preflightInterval = 10 * time.Minute
	scrapeTimeout     = 5 * time.Second
	// configFileMode is OBI's configuration's: readable by its user.
	configFileMode = 0o644
	// RowStatus is the "hp" of the node's status row.
	RowStatus = "obi"
)

// errScrapeStatus is OBI answering its metrics with an error.
var errScrapeStatus = errors.New("obi answered its metrics with an error")

// UC runs OBI on this node while it should, and writes its rows.
type UC struct {
	logger        logging.Logger
	db            database.IDB
	settingRepo   repository.SettingRepo
	appRepo       repository.AppRepo
	dockerManager docker.Manager

	// root is where the node's filesystem is: /host in the agent's container.
	root       string
	out        io.Writer
	httpClient *http.Client
	scrapeURL  string
	now        func() time.Time

	mu          sync.Mutex
	deltas      *obi.Deltas
	nodeID      string
	agentID     string
	logsOn      bool
	wanted      bool
	running     bool
	preflight   obi.Preflight
	preflightAt time.Time
	// capacity is the node's capacity as the settings choose it, auto for the
	// recommended one: the preflight is checked again when it changes.
	capacity obi.Capacity
	// appIDs are the opted-in apps, by their swarm service's name: what a
	// container's name starts with.
	appIDs map[string]string
}

func New(
	logger logging.Logger,
	db *database.DB,
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
	dockerManager docker.Manager,
	root string,
) *UC {
	return &UC{logger: logger, db: db, settingRepo: settingRepo, appRepo: appRepo, dockerManager: dockerManager,
		root: root, out: os.Stdout, httpClient: &http.Client{Timeout: scrapeTimeout},
		scrapeURL: fmt.Sprintf("http://127.0.0.1:%d/metrics", obi.MetricsPort), now: time.Now,
		deltas: obi.NewDeltas(), appIDs: map[string]string{}}
}

// Run reconciles at once and then on a timer, scrapes on a shorter one, and
// says the node's status on a longer one, until ctx ends.
func (uc *UC) Run(ctx context.Context) {
	reconcile := time.NewTicker(reconcileInterval)
	defer reconcile.Stop()
	scrape := time.NewTicker(scrapeInterval)
	defer scrape.Stop()
	status := time.NewTicker(statusInterval)
	defer status.Stop()
	uc.reconcileAndLog(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-reconcile.C:
			uc.reconcileAndLog(ctx)
		case <-scrape.C:
			if err := uc.Scrape(ctx); err != nil {
				uc.logger.Warnf("obi: scrape: %v", err)
			}
		case <-status.C:
			uc.writeStatus()
		}
	}
}

func (uc *UC) reconcileAndLog(ctx context.Context) {
	if err := uc.Reconcile(ctx); err != nil {
		uc.logger.Warnf("obi: reconcile: %v", err)
	}
}

// Reconcile reads the settings and makes OBI match them: running, with the
// opted-in apps, when this node runs it and can; gone otherwise. When the
// settings cannot be read, nothing changes.
func (uc *UC) Reconcile(ctx context.Context) error {
	cfg, err := uc.loggingSettings(ctx)
	if err != nil {
		return err
	}
	if uc.nodeID == "" {
		if uc.nodeID, err = uc.dockerManager.NodeCurrentID(ctx); err != nil {
			return hperrors.Wrap(err)
		}
	}
	logsOn := cfg.Enabled && (cfg.Sources.Apps || cfg.Sources.HivePaaS)
	node := cfg.Performance.Node(uc.nodeID)
	wanted := logsOn && node != nil
	capacity := obi.CapacityAuto
	if node != nil {
		if c, ok := obi.ParseCapacity(node.Capacity); ok {
			capacity = c
		}
	}
	appIDs := map[string]string{}
	if wanted {
		if appIDs, err = uc.optedInApps(ctx); err != nil {
			return err
		}
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.logsOn, uc.wanted, uc.appIDs = logsOn, wanted, appIDs
	if !uc.running {
		// One a previous agent left running: its memory is in the node's
		// free memory, which preflight must not count against it.
		uc.running = uc.containerRunning(ctx)
	}
	if uc.preflightAt.IsZero() || uc.now().Sub(uc.preflightAt) >= preflightInterval || capacity != uc.capacity {
		uc.preflight, uc.preflightAt, uc.capacity = obi.Check(uc.root, uc.running, capacity), uc.now(), capacity
	}
	if !wanted || !uc.preflight.OK || len(appIDs) == 0 {
		return uc.remove(ctx)
	}
	services := make([]string, 0, len(appIDs))
	for service := range appIDs {
		services = append(services, service)
	}
	return uc.ensure(ctx, obi.Config(obi.Patterns(services), uc.preflight.Capacity))
}

// loggingSettings are the logging settings, the defaults when there are none.
func (uc *UC) loggingSettings(ctx context.Context) (*entity.LoggingSettings, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, entity.NewObjectScopeGlobal(), base.SettingTypeLogging, true)
	if errors.Is(err, hperrors.ErrNotFound) || (err == nil && setting == nil) {
		return &entity.LoggingSettings{}, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	cfg, err := setting.AsLoggingSettings()
	return cfg, hperrors.Wrap(err)
}

// optedInApps are the apps asking for their routes and calls, by their swarm
// service's name.
func (uc *UC) optedInApps(ctx context.Context) (map[string]string, error) {
	settings, _, err := uc.settingRepo.List(ctx, uc.db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppFeatures),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhere("setting.data->'performanceSettings'->>'enabled' = 'true'"),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	ids := make([]string, 0, len(settings))
	for _, s := range settings {
		ids = append(ids, s.ObjectID)
	}
	out := make(map[string]string, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	apps, err := uc.appRepo.ListByIDs(ctx, uc.db, "", ids)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, app := range apps {
		if app.GlobalKey != "" && app.Status == base.AppStatusActive {
			out[app.GlobalKey] = app.ID
		}
	}
	return out, nil
}

// ensure runs OBI with a configuration: the one running is kept when it was
// made with it, in this agent's network namespace; otherwise it is replaced.
func (uc *UC) ensure(ctx context.Context, config []byte) error {
	hash := obi.ConfigHash(config)
	if uc.agentID == "" {
		id, err := uc.selfContainerID(ctx)
		if err != nil {
			return err
		}
		uc.agentID = id
	}
	current, err := uc.dockerManager.ContainerInspect(ctx, obi.ContainerName)
	if err != nil && !errors.Is(err, hperrors.ErrInfraNotFound) {
		return hperrors.Wrap(err)
	}
	if err == nil {
		c := current.Container
		if c.State != nil && c.State.Running && c.Config != nil && c.Config.Labels[obi.LabelConfig] == hash &&
			c.HostConfig != nil && string(c.HostConfig.NetworkMode) == "container:"+uc.agentID {
			uc.running = true
			return nil
		}
		if err = uc.removeContainer(ctx); err != nil {
			return err
		}
	}
	if err = uc.ensureImage(ctx); err != nil {
		return err
	}
	created, err := uc.dockerManager.ContainerCreate(ctx, func(opts *client.ContainerCreateOptions) {
		*opts = obi.ContainerOptions(uc.agentID, hash)
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	archive, err := tarOf(obi.ConfigFile, config)
	if err != nil {
		return err
	}
	if _, err = uc.dockerManager.ContainerCopyTo(ctx, created.ID, obi.ConfigDir, archive); err != nil {
		_ = uc.removeContainer(ctx)
		return hperrors.Wrap(err)
	}
	if _, err = uc.dockerManager.ContainerStart(ctx, created.ID); err != nil {
		_ = uc.removeContainer(ctx)
		return hperrors.Wrap(err)
	}
	// A new OBI counts from zero: its first scrape is a baseline.
	uc.deltas.Reset()
	uc.running = true
	uc.logger.Infof("obi: running for %d apps", len(uc.appIDs))
	return nil
}

// remove stops OBI when it runs.
func (uc *UC) remove(ctx context.Context) error {
	if !uc.running {
		_, err := uc.dockerManager.ContainerInspect(ctx, obi.ContainerName)
		if errors.Is(err, hperrors.ErrInfraNotFound) {
			return nil
		}
	}
	if err := uc.removeContainer(ctx); err != nil {
		return err
	}
	uc.running = false
	uc.deltas.Reset()
	return nil
}

// containerRunning is whether OBI's container runs, whoever started it.
func (uc *UC) containerRunning(ctx context.Context) bool {
	current, err := uc.dockerManager.ContainerInspect(ctx, obi.ContainerName)
	return err == nil && current.Container.State != nil && current.Container.State.Running
}

func (uc *UC) removeContainer(ctx context.Context) error {
	_, err := uc.dockerManager.ContainerRemove(ctx, obi.ContainerName, func(opts *client.ContainerRemoveOptions) {
		opts.Force = true
	})
	if err != nil && !errors.Is(err, hperrors.ErrInfraNotFound) {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) ensureImage(ctx context.Context) error {
	_, err := uc.dockerManager.ImageInspect(ctx, obi.Image)
	if err == nil {
		return nil
	}
	if !errors.Is(err, hperrors.ErrInfraNotFound) {
		return hperrors.Wrap(err)
	}
	pull, err := uc.dockerManager.ImagePull(ctx, obi.Image)
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() { _ = pull.Close() }()
	return hperrors.Wrap(pull.Wait(ctx))
}

// containerIDInMounts finds a container's id in its mount table: Docker mounts
// its hostname, hosts and resolv.conf from the container's own directory.
var containerIDInMounts = regexp.MustCompile(`/containers/([0-9a-f]{64})/`)

// selfContainerID is the agent's own container: OBI joins its network
// namespace.
func (uc *UC) selfContainerID(ctx context.Context) (string, error) {
	if mounts, err := os.ReadFile("/proc/self/mountinfo"); err == nil {
		if m := containerIDInMounts.FindSubmatch(mounts); m != nil {
			return string(m[1]), nil
		}
	}
	hostname, err := os.Hostname()
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	inspect, err := uc.dockerManager.ContainerInspect(ctx, hostname)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	return inspect.Container.ID, nil
}

// Scrape reads OBI's metrics and writes what moved since the last scrape, a
// row a series, for the containers of opted-in apps.
func (uc *UC) Scrape(ctx context.Context) error {
	uc.mu.Lock()
	running, appIDs := uc.running, uc.appIDs
	uc.mu.Unlock()
	if !running || len(appIDs) == 0 {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uc.scrapeURL, nil)
	if err != nil {
		return hperrors.Wrap(err)
	}
	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return hperrors.Wrap(err) // starting, or restarted: the next scrape
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: %d", errScrapeStatus, resp.StatusCode)
	}
	uc.mu.Lock()
	rows, err := uc.deltas.Read(resp.Body, func(name string) (string, bool) {
		return appOf(appIDs, name)
	})
	uc.mu.Unlock()
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(obi.WriteRows(uc.out, rows))
}

// appOf is the app a container is a task of: its name is its service's, then
// a dot.
func appOf(appIDs map[string]string, container string) (string, bool) {
	service, _, ok := strings.Cut(strings.TrimPrefix(container, "/"), ".")
	if !ok {
		return "", false
	}
	id, ok := appIDs[service]
	return id, ok
}

// Status is what a node can run and runs, as its status row says it.
type Status struct {
	HP        string        `json:"hp"`
	Node      string        `json:"node"`
	Wanted    bool          `json:"wanted"`
	Running   bool          `json:"running"`
	Apps      int           `json:"apps"`
	Preflight obi.Preflight `json:"preflight"`
}

// Status says what this node can run and runs.
func (uc *UC) Status() Status {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return Status{HP: RowStatus, Node: uc.nodeID, Wanted: uc.wanted, Running: uc.running, Apps: len(uc.appIDs),
		Preflight: uc.preflight}
}

// writeStatus writes the node's status row while the logs are stored: the
// settings show each node's from the latest. Before the first reconcile there
// is nothing to say.
func (uc *UC) writeStatus() {
	uc.mu.Lock()
	logsOn, known := uc.logsOn, !uc.preflightAt.IsZero()
	uc.mu.Unlock()
	if !logsOn || !known {
		return
	}
	line, err := json.Marshal(uc.Status())
	if err != nil {
		return
	}
	_, _ = uc.out.Write(append(line, '\n'))
}

// tarOf is one file in a tar archive, as ContainerCopyTo takes it.
func tarOf(name string, content []byte) (io.Reader, error) {
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)
	if err := w.WriteHeader(&tar.Header{Name: name, Mode: configFileMode, Size: int64(len(content))}); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if _, err := w.Write(content); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := w.Close(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &buf, nil
}
