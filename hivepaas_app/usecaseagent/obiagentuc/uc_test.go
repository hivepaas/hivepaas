package obiagentuc

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"iter"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/jsonstream"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/services/docker"
)

// settings answers the logging settings and the app features asked for, and
// counts the reads of the logging settings.
type settings struct {
	repository.SettingRepo
	logging  *entity.LoggingSettings
	features []*entity.Setting
	reads    int
}

func (f *settings) GetSingle(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	f.reads++
	if f.logging == nil {
		return nil, hperrors.NewNotFound("Setting")
	}
	s := &entity.Setting{Type: base.SettingTypeLogging, Status: base.SettingStatusActive}
	s.MustSetData(f.logging)
	return s, nil
}

func (f *settings) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	opts ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	var rows []*entity.Setting
	sql := bunex.ApplySelect(bun.NewDB(nil, pgdialect.New()).NewSelect().Model(&rows), opts...).String()
	if strings.Contains(sql, "performanceSettings") {
		return f.features, nil, nil
	}
	return nil, nil, nil
}

type apps struct {
	repository.AppRepo
	list []*entity.App
}

func (f *apps) ListByIDs(_ context.Context, _ database.IDB, _ string, ids []string,
	_ ...bunex.SelectQueryOption) ([]*entity.App, error) {
	var out []*entity.App
	for _, app := range f.list {
		for _, id := range ids {
			if app.ID == id {
				out = append(out, app)
			}
		}
	}
	return out, nil
}

// fakeDocker keeps one OBI container, and what was done to it; calls counts
// every call made to it.
type fakeDocker struct {
	docker.Manager
	obi     *container.InspectResponse
	config  string
	pulled  []string
	created int
	removed int
	calls   int
	// events are what was done, in order: pull, remove, create, start.
	events []string
}

func (f *fakeDocker) NodeCurrentID(context.Context) (string, error) {
	f.calls++
	return "node-1", nil
}

func (f *fakeDocker) ContainerInspect(_ context.Context, id string, _ ...docker.ContainerInspectOption) (
	*client.ContainerInspectResult, error) {
	f.calls++
	if id == obi.ContainerName && f.obi != nil {
		return &client.ContainerInspectResult{Container: *f.obi}, nil
	}
	if id != obi.ContainerName {
		return &client.ContainerInspectResult{Container: container.InspectResponse{ID: "agent-" + id}}, nil
	}
	return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
}

func (f *fakeDocker) ContainerRemove(_ context.Context, _ string, _ ...docker.ContainerRemoveOption) (
	*client.ContainerRemoveResult, error) {
	f.calls++
	if f.obi == nil {
		return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
	}
	f.obi = nil
	f.removed++
	f.events = append(f.events, "remove")
	return &client.ContainerRemoveResult{}, nil
}

// imageID is an image's id as the fake names it.
func imageID(ref string) string { return "id:" + ref }

func (f *fakeDocker) ImageInspect(_ context.Context, ref string, _ ...docker.ImageInspectOption) (
	*client.ImageInspectResult, error) {
	for _, p := range f.pulled {
		if p == ref {
			return &client.ImageInspectResult{InspectResponse: image.InspectResponse{ID: imageID(ref)}}, nil
		}
	}
	return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
}

type pullResponse struct{ io.ReadCloser }

func (pullResponse) Wait(context.Context) error { return nil }

func (pullResponse) JSONMessages(context.Context) iter.Seq2[jsonstream.Message, error] {
	return func(func(jsonstream.Message, error) bool) {}
}

func (f *fakeDocker) ImagePull(_ context.Context, image string, _ ...docker.ImagePullOption) (
	client.ImagePullResponse, error) {
	f.pulled = append(f.pulled, image)
	f.events = append(f.events, "pull")
	return pullResponse{ReadCloser: io.NopCloser(&bytes.Buffer{})}, nil
}

func (f *fakeDocker) ContainerCreate(_ context.Context, options ...docker.ContainerCreateOption) (
	*client.ContainerCreateResult, error) {
	opts := &client.ContainerCreateOptions{}
	for _, o := range options {
		o(opts)
	}
	f.obi = &container.InspectResponse{ID: "obi-1", Image: imageID(opts.Config.Image), Config: opts.Config,
		HostConfig: opts.HostConfig, State: &container.State{}}
	f.created++
	f.events = append(f.events, "create")
	return &client.ContainerCreateResult{ID: "obi-1"}, nil
}

func (f *fakeDocker) ContainerCopyTo(_ context.Context, _ string, dst string, content io.Reader,
	_ ...docker.ContainerCopyToOption) (*client.CopyToContainerResult, error) {
	r := tar.NewReader(content)
	h, err := r.Next()
	if err != nil {
		return nil, err
	}
	b, _ := io.ReadAll(r)
	f.config = dst + "/" + h.Name + ":\n" + string(b)
	return &client.CopyToContainerResult{}, nil
}

func (f *fakeDocker) ContainerStart(_ context.Context, _ string, _ ...docker.ContainerStartOption) (
	*client.ContainerStartResult, error) {
	f.obi.State.Running = true
	f.events = append(f.events, "start")
	return &client.ContainerStartResult{}, nil
}

type quiet struct{ logging.Logger }

func (quiet) Infof(string, ...any) {}
func (quiet) Warnf(string, ...any) {}

// goodNode is a node's filesystem OBI can run on.
func goodNode(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for name, content := range map[string]string{
		"proc/sys/kernel/osrelease": "6.8.0-124-generic\n",
		"sys/kernel/btf/vmlinux":    "btf",
		"sys/kernel/tracing/trace":  "",
		"proc/meminfo":              "MemAvailable: 466944 kB\n",
	} {
		path := filepath.Join(root, name)
		assert.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
		assert.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
	return root
}

func featuresOf(appID string) *entity.Setting {
	s := &entity.Setting{Type: base.SettingTypeAppFeatures, ObjectID: appID, Status: base.SettingStatusActive}
	s.MustSetData(&entity.AppFeatureSettings{PerformanceSettings: &entity.AppFeaturePerformanceSettings{Enabled: true}})
	return s
}

type world struct {
	uc       *UC
	settings *settings
	docker   *fakeDocker
	out      *bytes.Buffer
}

func nodesOf(ids ...string) []*entity.LoggingPerformanceNode {
	out := make([]*entity.LoggingPerformanceNode, 0, len(ids))
	for _, id := range ids {
		out = append(out, &entity.LoggingPerformanceNode{ID: id})
	}
	return out
}

func newWorld(t *testing.T, nodes ...string) *world {
	t.Helper()
	w := &world{
		settings: &settings{
			logging: &entity.LoggingSettings{Enabled: true, Sources: entity.LoggingSources{Apps: true},
				Performance: &entity.LoggingPerformance{Enabled: true, Nodes: nodesOf(nodes...)}},
			features: []*entity.Setting{featuresOf("A2")},
		},
		docker: &fakeDocker{},
		out:    &bytes.Buffer{},
	}
	w.uc = New(quiet{}, nil, w.settings, &apps{list: []*entity.App{
		{ID: "A1", GlobalKey: "p1_dev_a1", Status: base.AppStatusActive},
		{ID: "A2", GlobalKey: "p1_dev_a2", Status: base.AppStatusActive},
	}}, nil, w.docker, goodNode(t))
	w.uc.out = w.out
	w.uc.agentID = "agent-container"
	return w
}

// A node the settings list runs OBI for the opted-in apps, made once: the
// image pulled, its configuration copied in, in the agent's network.
func TestReconcileRunsOBIOnAListedNode(t *testing.T) {
	w := newWorld(t, "node-1")
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 1, w.docker.created)
	assert.Equal(t, []string{w.uc.image}, w.docker.pulled)
	assert.Contains(t, w.docker.config, "//hivepaas-obi.yaml:")
	assert.Contains(t, w.docker.config, `container_name: "p1_dev_a2.*"`)
	assert.Contains(t, w.docker.config, "global_scale_factor: -2", "the recommended capacity: small, on 456 MB")
	assert.NotContains(t, w.docker.config, "p1_dev_a1", "not opted in")
	assert.Equal(t, "container:agent-container", string(w.docker.obi.HostConfig.NetworkMode))

	// The same: kept as it is.
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 1, w.docker.created)
	assert.Zero(t, w.docker.removed)

	// Another app opts in: replaced, watching both.
	w.settings.features = append(w.settings.features, featuresOf("A1"))
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 2, w.docker.created)
	assert.Equal(t, 1, w.docker.removed)
	assert.Contains(t, w.docker.config, `container_name: "p1_dev_a1.*"`)
}

// OBI is removed when the node is not listed, the feature is off, the logs
// are not stored, no app asks, or the node cannot run it.
func TestReconcileRemovesOBI(t *testing.T) {
	cases := map[string]func(w *world){
		"node not listed": func(w *world) { w.settings.logging.Performance.Nodes = nodesOf("node-2") },
		"feature off":     func(w *world) { w.settings.logging.Performance.Enabled = false },
		"logs not stored": func(w *world) { w.settings.logging.Enabled = false },
		"no app asks":     func(w *world) { w.settings.features = nil },
		"no btf": func(w *world) {
			assert.NoError(t, os.Remove(filepath.Join(w.uc.root, "sys/kernel/btf/vmlinux")))
			w.uc.preflightAt = time.Time{}
		},
	}
	for name, change := range cases {
		t.Run(name, func(t *testing.T) {
			w := newWorld(t, "node-1")
			assert.NoError(t, w.uc.Reconcile(context.Background()))
			assert.True(t, w.uc.running)
			change(w)
			assert.NoError(t, w.uc.Reconcile(context.Background()))
			assert.Nil(t, w.docker.obi, "removed")
			assert.False(t, w.uc.running)
		})
	}
}

// One left running by a previous agent, in another agent's network: replaced.
func TestReconcileReplacesOBIOfAnotherAgent(t *testing.T) {
	w := newWorld(t, "node-1")
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	w.docker.obi.HostConfig.NetworkMode = "container:old-agent"
	w.uc.checked, w.uc.running = false, false // a new agent
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 2, w.docker.created)
	assert.Equal(t, "container:agent-container", string(w.docker.obi.HostConfig.NetworkMode))
}

// Scraping writes, after a baseline, a row per series that moved, for the
// opted-in apps' containers.
func TestScrapeWritesRows(t *testing.T) {
	count := 100
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(rw, `http_server_request_duration_seconds_bucket{container_name="p1_dev_a2.1.abc",`+
			`http_request_method="GET",http_response_status_code="200",http_route="/u/*",le="+Inf"} `+
			itoa(count)+"\n"+
			`http_server_request_duration_seconds_sum{container_name="p1_dev_a2.1.abc",http_request_method="GET",`+
			`http_response_status_code="200",http_route="/u/*"} 1`+"\n"+
			`http_server_request_duration_seconds_count{container_name="p1_dev_a2.1.abc",http_request_method="GET",`+
			`http_response_status_code="200",http_route="/u/*"} `+itoa(count)+"\n"+
			`http_server_request_duration_seconds_count{container_name="stranger.1.x",http_route="/"} `+
			itoa(count)+"\n")
	}))
	defer srv.Close()

	w := newWorld(t, "node-1")
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	w.uc.scrapeURL = srv.URL
	assert.NoError(t, w.uc.Scrape(context.Background()))
	assert.Empty(t, w.out.String(), "a baseline")

	count = 130
	assert.NoError(t, w.uc.Scrape(context.Background()))
	lines := strings.Split(strings.TrimSpace(w.out.String()), "\n")
	if assert.Len(t, lines, 1) {
		var row map[string]any
		assert.NoError(t, json.Unmarshal([]byte(lines[0]), &row))
		assert.Equal(t, "routes", row["hp"])
		assert.Equal(t, "A2", row["app"])
		assert.InDelta(t, 30, row["count"], 0)
		assert.InDelta(t, 30, row["leInf"], 0)
	}
}

// The status row says what the node can run and runs, while logs are stored.
func TestStatusRow(t *testing.T) {
	w := newWorld(t, "node-1")
	w.uc.writeStatus()
	assert.Empty(t, w.out.String(), "nothing known before the first reconcile")

	assert.NoError(t, w.uc.Reconcile(context.Background()))
	w.uc.writeStatus()
	var st obi.Status
	assert.NoError(t, json.Unmarshal(w.out.Bytes(), &st))
	assert.Equal(t, obi.Status{HP: "obi", Node: "node-1", Wanted: true, Running: true, Apps: 1,
		Preflight: obi.Preflight{OK: true, Kernel: "6.8.0-124-generic", MemAvailableMB: 456,
			Recommended: obi.CapacitySmall, Capacity: obi.CapacitySmall}}, st)

	w.out.Reset()
	w.settings.logging.Enabled = false
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	w.uc.writeStatus()
	assert.Empty(t, w.out.String(), "no logs stored: nobody reads it")
}

func TestAppOf(t *testing.T) {
	ids := map[string]string{"p1_dev_a2": "A2"}
	id, ok := appOf(ids, "/p1_dev_a2.1.abc")
	assert.True(t, ok)
	assert.Equal(t, "A2", id)
	_, ok = appOf(ids, "p1_dev_a20.1.abc")
	assert.False(t, ok)
	_, ok = appOf(ids, "p1_dev_a2")
	assert.False(t, ok)
}

func itoa(n int) string { return strconv.Itoa(n) }

// A capacity chosen for the node sizes OBI's maps; changing it replaces OBI.
func TestReconcileSizesOBIByTheNodesCapacity(t *testing.T) {
	w := newWorld(t, "node-1")
	w.settings.logging.Performance.Nodes[0].Capacity = "medium"
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Contains(t, w.docker.config, "global_scale_factor: -1")
	assert.Equal(t, obi.CapacityMedium, w.uc.Status().Preflight.Capacity)
	assert.Equal(t, obi.CapacitySmall, w.uc.Status().Preflight.Recommended)

	w.settings.logging.Performance.Nodes[0].Capacity = "auto"
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 2, w.docker.created, "replaced")
	assert.Contains(t, w.docker.config, "global_scale_factor: -2")
}

// On a node whose kernel restricts perf events, OBI is given CAP_SYS_ADMIN, or
// it attaches no probe; once they are restricted no more, it is replaced
// without.
func TestReconcileGivesOBISysAdminWherePerfEventsAreRestricted(t *testing.T) {
	w := newWorld(t, "node-1")
	paranoid := filepath.Join(w.uc.root, "proc/sys/kernel/perf_event_paranoid")
	assert.NoError(t, os.WriteFile(paranoid, []byte("3\n"), 0o600))
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Contains(t, w.docker.obi.HostConfig.CapAdd, "CAP_SYS_ADMIN")

	assert.NoError(t, os.WriteFile(paranoid, []byte("2\n"), 0o600))
	w.uc.preflightAt = time.Time{}
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 2, w.docker.created, "replaced")
	assert.NotContains(t, w.docker.obi.HostConfig.CapAdd, "CAP_SYS_ADMIN")
}

// clock is a time a test moves.
type clock struct{ at time.Time }

func (c *clock) now() time.Time { return c.at }

// While the feature is off, the agent reads the settings and, every 10
// minutes, looks at the node as when it is on: a stray OBI goes, the node says
// what it can run. In between, no Docker, no node's files, no rows, no timers.
func TestReconcileWhileOffLooksEveryTenMinutes(t *testing.T) {
	w := newWorld(t, "node-1")
	c := &clock{at: time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)}
	w.uc.now = c.now
	w.settings.logging.Performance.Enabled = false
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 2, w.docker.calls, "a look for a leftover, the node's id")
	assert.False(t, w.uc.preflightAt.IsZero(), "the node checked")
	var st obi.Status
	assert.NoError(t, json.Unmarshal(w.out.Bytes(), &st), "it says what it can run")
	assert.True(t, st.Preflight.OK)
	assert.False(t, st.Wanted)

	w.out.Reset()
	for range 3 {
		c.at = c.at.Add(30 * time.Second)
		assert.NoError(t, w.uc.Reconcile(context.Background()))
	}
	assert.Equal(t, 2, w.docker.calls, "nothing in between")
	w.uc.writeStatus()
	assert.Empty(t, w.out.String(), "no row in between")
	scrape, status := w.uc.activity()
	assert.False(t, scrape)
	assert.False(t, status)

	// One started since, by hand or by an agent from before: gone at the next
	// look, and the node says its status again.
	w.docker.obi = &container.InspectResponse{ID: "stray", State: &container.State{Running: true},
		HostConfig: &container.HostConfig{}, Config: &container.Config{}}
	c.at = c.at.Add(10 * time.Minute)
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Nil(t, w.docker.obi)
	assert.Equal(t, 3, w.docker.calls)
	assert.Contains(t, w.out.String(), `"hp":"obi"`)

	// Turned on: the node is checked, OBI made, and the timers needed.
	w.settings.logging.Performance.Enabled = true
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, 1, w.docker.created)
	assert.False(t, w.uc.preflightAt.IsZero())
	scrape, status = w.uc.activity()
	assert.True(t, scrape)
	assert.True(t, status)

	// Off again: removed once, then nothing.
	w.settings.logging.Performance.Enabled = false
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Nil(t, w.docker.obi)
	calls := w.docker.calls
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, calls, w.docker.calls)
}

// A node the settings do not list, while the feature is on, says its status
// and asks Docker nothing once it has looked for a leftover.
func TestReconcileOnAnotherNodeAsksDockerOnce(t *testing.T) {
	w := newWorld(t, "node-2")
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	calls := w.docker.calls
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, calls, w.docker.calls)
	assert.Zero(t, w.docker.created)
	w.uc.writeStatus()
	assert.Contains(t, w.out.String(), `"wanted":false`)
	scrape, status := w.uc.activity()
	assert.False(t, scrape, "OBI does not run here")
	assert.True(t, status)
}

// A stopped OBI a previous agent left is removed, and none is made where it
// is not wanted.
func TestReconcileRemovesAStoppedLeftover(t *testing.T) {
	w := newWorld(t, "node-2")
	w.docker.obi = &container.InspectResponse{ID: "old", State: &container.State{Running: false},
		HostConfig: &container.HostConfig{}, Config: &container.Config{}}
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Nil(t, w.docker.obi)
	assert.Equal(t, 1, w.docker.removed)
}

// A ticker is stopped when it is not needed and started again when it is.
func TestPace(t *testing.T) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	assert.False(t, pace(ticker, time.Millisecond, true, false))
	assert.True(t, pace(ticker, time.Millisecond, false, true))
	select {
	case <-ticker.C:
	case <-time.After(time.Second):
		t.Fatal("started again: it ticks")
	}
}

// cache answers what it was given, at a generation, and keeps what is set.
type cache struct {
	settings *entity.OBIAgentSettings
	gen      int64
	err      error
	set      []int64
}

func (c *cache) Get(context.Context) (*entity.OBIAgentSettings, int64, error) {
	return c.settings, c.gen, c.err
}

func (c *cache) Set(_ context.Context, s *entity.OBIAgentSettings, gen int64, _ time.Duration) error {
	c.settings, c.set = s, append(c.set, gen)
	return nil
}

func (c *cache) Invalidate(context.Context) error { return nil }

// The settings come from the cache while it holds them; from the database on
// a miss and every 10 minutes - then cached under the generation read before -
// and when Redis cannot be read.
func TestSettingsComeFromTheCache(t *testing.T) {
	w := newWorld(t, "node-1")
	c := &clock{at: time.Date(2026, 10, 4, 10, 0, 0, 0, time.UTC)}
	cached := &cache{gen: 7}
	w.uc.now, w.uc.cache = c.now, cached

	s, err := w.uc.settings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, w.settings.reads, "nothing cached: the database")
	assert.Equal(t, []int64{7}, cached.set, "cached under the generation read before")
	assert.Equal(t, map[string]string{"p1_dev_a2": "A2"}, s.Apps, "the opted-in apps with them, the feature on")

	for range 5 {
		c.at = c.at.Add(30 * time.Second)
		_, err = w.uc.settings(context.Background())
		assert.NoError(t, err)
	}
	assert.Equal(t, 1, w.settings.reads, "from the cache")

	c.at = c.at.Add(10 * time.Minute)
	_, err = w.uc.settings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, w.settings.reads, "every 10 minutes, the database")

	// A change: the entry dropped, the next read is the database's.
	cached.settings, cached.gen = nil, 8
	_, err = w.uc.settings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 3, w.settings.reads)
	assert.Equal(t, int64(8), cached.set[len(cached.set)-1])

	// Redis down: the database, nothing cached.
	cached.settings, cached.err = nil, errors.New("redis down")
	sets := len(cached.set)
	_, err = w.uc.settings(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 4, w.settings.reads)
	assert.Len(t, cached.set, sets)

	// While the feature is off, no app is read.
	w.settings.logging.Performance.Enabled = false
	cached.err = nil
	c.at = c.at.Add(10 * time.Minute)
	s, err = w.uc.settings(context.Background())
	assert.NoError(t, err)
	assert.False(t, s.On())
	assert.Nil(t, s.Apps)
}

// The OBI an agent runs is its release's. A new release's - the agent updated
// - is pulled while the old one still runs, then swapped; the old image is
// left to the system cleanup.
func TestANewReleasesOBIIsPulledBeforeTheSwap(t *testing.T) {
	w := newWorld(t, "node-1")
	assert.Equal(t, base.BetaVersion.OBIImage, w.uc.image, "the release's, as it was built")
	w.uc.image = "otel/ebpf-instrument:v0.14.0"
	assert.NoError(t, w.uc.Reconcile(context.Background()))

	w.docker.events = nil
	w.uc.image = "otel/ebpf-instrument:v0.15.0"
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, []string{"pull", "remove", "create", "start"}, w.docker.events)
	assert.Equal(t, "otel/ebpf-instrument:v0.15.0", w.docker.obi.Config.Image)

	// Another app asks: the image is there, nothing to pull.
	w.docker.events = nil
	w.settings.features = append(w.settings.features, featuresOf("A1"))
	assert.NoError(t, w.uc.Reconcile(context.Background()))
	assert.Equal(t, []string{"remove", "create", "start"}, w.docker.events)
}
