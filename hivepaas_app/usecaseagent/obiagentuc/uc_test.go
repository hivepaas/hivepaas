package obiagentuc

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
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

// settings answers the logging settings and the app features asked for.
type settings struct {
	repository.SettingRepo
	logging  *entity.LoggingSettings
	features []*entity.Setting
}

func (f *settings) GetSingle(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	_ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
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

// fakeDocker keeps one OBI container, and what was done to it.
type fakeDocker struct {
	docker.Manager
	obi     *container.InspectResponse
	config  string
	pulled  []string
	created int
	removed int
}

func (f *fakeDocker) NodeCurrentID(context.Context) (string, error) { return "node-1", nil }

func (f *fakeDocker) ContainerInspect(_ context.Context, id string, _ ...docker.ContainerInspectOption) (
	*client.ContainerInspectResult, error) {
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
	if f.obi == nil {
		return nil, hperrors.Wrap(hperrors.ErrInfraNotFound)
	}
	f.obi = nil
	f.removed++
	return &client.ContainerRemoveResult{}, nil
}

func (f *fakeDocker) ImageInspect(_ context.Context, image string, _ ...docker.ImageInspectOption) (
	*client.ImageInspectResult, error) {
	for _, p := range f.pulled {
		if p == image {
			return &client.ImageInspectResult{}, nil
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
	return pullResponse{ReadCloser: io.NopCloser(&bytes.Buffer{})}, nil
}

func (f *fakeDocker) ContainerCreate(_ context.Context, options ...docker.ContainerCreateOption) (
	*client.ContainerCreateResult, error) {
	opts := &client.ContainerCreateOptions{}
	for _, o := range options {
		o(opts)
	}
	f.obi = &container.InspectResponse{ID: "obi-1", Config: opts.Config, HostConfig: opts.HostConfig,
		State: &container.State{}}
	f.created++
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

func newWorld(t *testing.T, nodes ...string) *world {
	t.Helper()
	w := &world{
		settings: &settings{
			logging: &entity.LoggingSettings{Enabled: true, Sources: entity.LoggingSources{Apps: true},
				Performance: &entity.LoggingPerformance{Enabled: true, Nodes: nodes}},
			features: []*entity.Setting{featuresOf("A2")},
		},
		docker: &fakeDocker{},
		out:    &bytes.Buffer{},
	}
	w.uc = New(quiet{}, nil, w.settings, &apps{list: []*entity.App{
		{ID: "A1", GlobalKey: "p1_dev_a1", Status: base.AppStatusActive},
		{ID: "A2", GlobalKey: "p1_dev_a2", Status: base.AppStatusActive},
	}}, w.docker, goodNode(t))
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
	assert.Equal(t, []string{obi.Image}, w.docker.pulled)
	assert.Contains(t, w.docker.config, "//hivepaas-obi.yaml:")
	assert.Contains(t, w.docker.config, `container_name: "p1_dev_a2.*"`)
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
		"node not listed": func(w *world) { w.settings.logging.Performance.Nodes = []string{"node-2"} },
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
	w.uc.running = false
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
	var st Status
	assert.NoError(t, json.Unmarshal(w.out.Bytes(), &st))
	assert.Equal(t, Status{HP: "obi", Node: "node-1", Wanted: true, Running: true, Apps: 1,
		Preflight: obi.Preflight{OK: true, Kernel: "6.8.0-124-generic", MemAvailableMB: 456}}, st)

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
