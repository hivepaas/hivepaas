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
	"github.com/hivepaas/hivepaas/hivepaas_app/repository/cacherepository"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// reconcileInterval is how often the settings are read and OBI made to
	// match them: a switch turned takes this long at most.
	reconcileInterval = 30 * time.Second
	// scrapeInterval is how often OBI's metrics are read: a row per series
	// per interval, as the agent's resource rows.
	scrapeInterval = 15 * time.Second
	// statusInterval is how often the node says what it can run and runs,
	// while the feature is on.
	statusInterval = obi.StatusEvery
	// preflightInterval is how often the node is checked again: a kernel
	// does not change while it runs, its free memory does.
	preflightInterval = 10 * time.Minute
	// offInterval is how often, while the feature is off, the node is looked
	// at as when it is on: an OBI that should not run goes, and the node says
	// what it can run. In between, the agent reads the settings, nothing else.
	offInterval = obi.StatusEveryOff
	// syncInterval is how often the settings are read from the database
	// rather than the cache: what a missed change can cost at most.
	syncInterval = 10 * time.Minute
	// cacheTTL is how long settings read from the database are cached for the
	// agents.
	cacheTTL      = time.Hour
	scrapeTimeout = 5 * time.Second
	// configFileMode is OBI's configuration's: readable by its user.
	configFileMode = 0o644
)

// errScrapeStatus is OBI answering its metrics with an error.
var errScrapeStatus = errors.New("obi answered its metrics with an error")

// UC runs OBI on this node while it should, and writes its rows.
type UC struct {
	logger        logging.Logger
	db            database.IDB
	settingRepo   repository.SettingRepo
	appRepo       repository.AppRepo
	cache         cacherepository.OBISettingsRepo
	dockerManager docker.Manager

	// root is where the node's filesystem is: /host in the agent's container.
	root       string
	out        io.Writer
	httpClient *http.Client
	scrapeURL  string
	now        func() time.Time

	mu      sync.Mutex
	deltas  *obi.Deltas
	nodeID  string
	agentID string
	// on is the feature on while the logs are stored: until it is, the agent
	// reads the settings and does nothing else.
	on     bool
	wanted bool
	// checked says this agent has looked for an OBI container since it
	// started: one a previous agent left. From then on, running is what this
	// agent made it.
	checked     bool
	running     bool
	preflight   obi.Preflight
	preflightAt time.Time
	// capacity is the node's capacity as the settings choose it, auto for the
	// recommended one: the preflight is checked again when it changes.
	capacity obi.Capacity
	// appIDs are the opted-in apps, by their swarm service's name: what a
	// container's name starts with.
	appIDs map[string]string
	// syncedAt is when the settings were last read from the database;
	// offCheckedAt when the node was last looked at while the feature is off.
	syncedAt     time.Time
	offCheckedAt time.Time
}

func New(
	logger logging.Logger,
	db *database.DB,
	settingRepo repository.SettingRepo,
	appRepo repository.AppRepo,
	cache cacherepository.OBISettingsRepo,
	dockerManager docker.Manager,
	root string,
) *UC {
	return &UC{logger: logger, db: db, settingRepo: settingRepo, appRepo: appRepo, cache: cache,
		dockerManager: dockerManager,
		root:          root, out: os.Stdout, httpClient: &http.Client{Timeout: scrapeTimeout},
		scrapeURL: fmt.Sprintf("http://127.0.0.1:%d/metrics", obi.MetricsPort), now: time.Now,
		deltas: obi.NewDeltas(), appIDs: map[string]string{}}
}

// Run reconciles at once and then on a timer until ctx ends. While OBI runs
// here it scrapes it on a shorter timer, and while the feature is on it says
// the node's status on a longer one: otherwise those timers are stopped.
func (uc *UC) Run(ctx context.Context) {
	reconcile := time.NewTicker(reconcileInterval)
	defer reconcile.Stop()
	scrape := time.NewTicker(scrapeInterval)
	defer scrape.Stop()
	status := time.NewTicker(statusInterval)
	defer status.Stop()
	scraping, saying := true, true
	reconcileAndPace := func() {
		uc.reconcileAndLog(ctx)
		scrapeNow, sayNow := uc.activity()
		scraping = pace(scrape, scrapeInterval, scraping, scrapeNow)
		saying = pace(status, statusInterval, saying, sayNow)
	}
	reconcileAndPace()
	for {
		select {
		case <-ctx.Done():
			return
		case <-reconcile.C:
			reconcileAndPace()
		case <-scrape.C:
			if err := uc.Scrape(ctx); err != nil {
				uc.logger.Warnf("obi: scrape: %v", err)
			}
		case <-status.C:
			uc.writeStatus()
		}
	}
}

// pace starts or stops a ticker as it is now needed, and says whether it
// runs. A tick already sent before it stops is read once, and does nothing.
func pace(t *time.Ticker, every time.Duration, running, needed bool) bool {
	switch {
	case needed && !running:
		t.Reset(every)
	case !needed && running:
		t.Stop()
	}
	return needed
}

// activity is what the timers are needed for: scraping while OBI runs here,
// the status while the feature is on.
func (uc *UC) activity() (scrape, status bool) {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.running, uc.on
}

func (uc *UC) reconcileAndLog(ctx context.Context) {
	if err := uc.Reconcile(ctx); err != nil {
		uc.logger.Warnf("obi: reconcile: %v", err)
	}
}

// Reconcile reads the settings and makes OBI match them: running, with the
// opted-in apps, when this node runs it and can; gone otherwise. When the
// settings cannot be read, nothing changes. While the feature is off, reading
// them is all it does, but every offInterval.
func (uc *UC) Reconcile(ctx context.Context) error {
	s, err := uc.settings(ctx)
	if err != nil {
		return err
	}
	if !s.On() {
		return uc.off(ctx, s)
	}
	if err = uc.ensureNodeID(ctx); err != nil {
		return err
	}
	node := s.Performance.Node(uc.nodeID)
	wanted := node != nil
	capacity := capacityOf(node)
	appIDs := map[string]string{}
	if wanted {
		for service, id := range s.Apps {
			appIDs[service] = id
		}
	}

	uc.mu.Lock()
	defer uc.mu.Unlock()
	// Turned off again, the node is looked at at once.
	uc.on, uc.wanted, uc.appIDs, uc.offCheckedAt = true, wanted, appIDs, time.Time{}
	if !uc.checked {
		// One a previous agent left running: its memory is in the node's
		// free memory, which preflight must not count against it.
		if err = uc.adoptLeftover(ctx); err != nil {
			return err
		}
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

// off is the feature off, or the logs. The node is looked at as when it is on,
// at once and then every offInterval, the time kept in offCheckedAt: an OBI
// that should not run goes - one a previous agent left, this one ran, or one
// started since - and, while the logs are stored, the node says what it can
// run. In between nothing is done: no Docker, no files, no rows.
func (uc *UC) off(ctx context.Context, s *entity.OBIAgentSettings) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	uc.on, uc.wanted, uc.appIDs = false, false, map[string]string{}
	if !uc.offCheckedAt.IsZero() && uc.now().Sub(uc.offCheckedAt) < offInterval {
		return nil
	}
	uc.checked = false // look again: one may have been started since
	if err := uc.remove(ctx); err != nil {
		return err
	}
	if !s.LogsOn {
		// Nobody would read what the node says.
		uc.preflight, uc.preflightAt, uc.offCheckedAt = obi.Preflight{}, time.Time{}, uc.now()
		return nil
	}
	if err := uc.ensureNodeID(ctx); err != nil {
		return err
	}
	capacity := capacityOf(chosenNode(s.Performance, uc.nodeID))
	uc.preflight, uc.preflightAt, uc.capacity = obi.Check(uc.root, false, capacity), uc.now(), capacity
	uc.offCheckedAt = uc.now()
	uc.writeRow(uc.statusLocked())
	return nil
}

// ensureNodeID learns this node's id, once.
func (uc *UC) ensureNodeID(ctx context.Context) error {
	if uc.nodeID != "" {
		return nil
	}
	id, err := uc.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	uc.nodeID = id
	return nil
}

// chosenNode is the node's entry in the settings, the feature on or off: nil
// when it is not listed.
func chosenNode(perf *entity.LoggingPerformance, nodeID string) *entity.LoggingPerformanceNode {
	if perf == nil {
		return nil
	}
	for _, node := range perf.Nodes {
		if node != nil && node.ID == nodeID {
			return node
		}
	}
	return nil
}

// capacityOf is a node's capacity as chosen: auto when none is.
func capacityOf(node *entity.LoggingPerformanceNode) obi.Capacity {
	if node != nil {
		if c, ok := obi.ParseCapacity(node.Capacity); ok {
			return c
		}
	}
	return obi.CapacityAuto
}

// adoptLeftover looks once for an OBI a previous agent left: a running one is
// kept as running, to be kept or replaced; a stopped one is removed.
func (uc *UC) adoptLeftover(ctx context.Context) error {
	current, err := uc.dockerManager.ContainerInspect(ctx, obi.ContainerName)
	switch {
	case errors.Is(err, hperrors.ErrInfraNotFound):
		uc.running = false
	case err != nil:
		return hperrors.Wrap(err)
	case current.Container.State != nil && current.Container.State.Running:
		uc.running = true
	default:
		if err = uc.removeContainer(ctx); err != nil {
			return err
		}
		uc.running = false
	}
	uc.checked = true
	return nil
}

// settings are what this agent runs OBI by: from the cache while it holds
// the current ones; from the database every syncInterval, the time kept in
// syncedAt, and whenever the cache does not hold them - then cached for every
// agent, under the generation read before. Without Redis, the database.
func (uc *UC) settings(ctx context.Context) (*entity.OBIAgentSettings, error) {
	sync := uc.syncedAt.IsZero() || uc.now().Sub(uc.syncedAt) >= syncInterval
	var gen int64
	cacheOK := false
	if uc.cache != nil {
		cached, g, err := uc.cache.Get(ctx)
		if err == nil && cached != nil && !sync {
			return cached, nil
		}
		gen, cacheOK = g, err == nil
	}
	s, err := uc.readSettings(ctx)
	if err != nil {
		return nil, err
	}
	uc.syncedAt = uc.now()
	if cacheOK {
		if err = uc.cache.Set(ctx, s, gen, cacheTTL); err != nil {
			uc.logger.Warnf("obi: caching the settings: %v", err)
		}
	}
	return s, nil
}

// readSettings reads the settings from the database: the logging settings,
// and the opted-in apps while the feature is on.
func (uc *UC) readSettings(ctx context.Context) (*entity.OBIAgentSettings, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, uc.db, entity.NewObjectScopeGlobal(), base.SettingTypeLogging, true)
	if errors.Is(err, hperrors.ErrNotFound) || (err == nil && setting == nil) {
		return &entity.OBIAgentSettings{}, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	cfg, err := setting.AsLoggingSettings()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	s := &entity.OBIAgentSettings{LogsOn: cfg.Enabled && (cfg.Sources.Apps || cfg.Sources.HivePaaS),
		Performance: cfg.Performance}
	if s.On() {
		if s.Apps, err = uc.optedInApps(ctx); err != nil {
			return nil, err
		}
	}
	return s, nil
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

// remove takes OBI away: when it runs, or this agent has not looked for one
// yet. Once there is none, removing it again asks Docker nothing.
func (uc *UC) remove(ctx context.Context) error {
	if uc.checked && !uc.running {
		return nil
	}
	if err := uc.removeContainer(ctx); err != nil {
		return err
	}
	uc.checked, uc.running = true, false
	uc.deltas.Reset()
	return nil
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

// Status says what this node can run and runs.
func (uc *UC) Status() obi.Status {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	return uc.statusLocked()
}

func (uc *UC) statusLocked() obi.Status {
	return obi.Status{HP: obi.RowStatus, Node: uc.nodeID, Wanted: uc.wanted, Running: uc.running,
		Apps: len(uc.appIDs), Preflight: uc.preflight}
}

// writeStatus writes the node's status row while the feature is on, the
// logs stored: the settings show each node's from the latest. Before a
// reconcile has checked the node there is nothing to say. While the feature
// is off, off writes it, every offInterval.
func (uc *UC) writeStatus() {
	uc.mu.Lock()
	on, known, status := uc.on, !uc.preflightAt.IsZero(), uc.statusLocked()
	uc.mu.Unlock()
	if !on || !known {
		return
	}
	uc.writeRow(status)
}

// writeRow writes a status row to the agent's stdout.
func (uc *UC) writeRow(status obi.Status) {
	line, err := json.Marshal(status)
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
