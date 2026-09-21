package apptemplateserviceimpl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

type recordingLogger struct {
	logging.Logger
	warnings []string
	infos    []string
}

func (l *recordingLogger) Warnf(template string, _ ...any) {
	l.warnings = append(l.warnings, template)
}

func (l *recordingLogger) Infof(template string, _ ...any) {
	l.infos = append(l.infos, template)
}

// unavailableSource is the official source when GitHub cannot be reached.
type unavailableSource struct{}

func (unavailableSource) ID() string { return officialSourceID }
func (unavailableSource) Revision(context.Context) (string, error) {
	return "", hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}
func (unavailableSource) Index(context.Context) (*templatemodel.Index, error) {
	return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}
func (unavailableSource) TemplateFile(context.Context, *templatemodel.IndexEntry) ([]byte, error) {
	return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}
func (unavailableSource) Icon(context.Context, string) ([]byte, error) {
	return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable)
}

func newServiceTest(t *testing.T, env, dir string) (*service, *recordingLogger) {
	t.Helper()
	previous := config.Current()
	config.SetCurrent(&config.Config{Env: env, AppTemplates: config.AppTemplates{Dir: dir}})
	t.Cleanup(func() { config.SetCurrent(previous) })

	logger := &recordingLogger{}
	return &service{official: unavailableSource{}, logger: logger}, logger
}

// copyTestRepo copies testdata/repo so a test can change a template.
func copyTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	assert.NoError(t, os.CopyFS(dir, os.DirFS(testRepoDir)))
	return dir
}

func TestServiceReadsTheLocalDirectoryInDevelopment(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	index, err := svc.Index(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, localSourceID, index.Source)
	assert.NotNil(t, index.Index.FindTemplate("demo"))
}

func TestServiceIgnoresTheLocalDirectoryOutsideDevelopment(t *testing.T) {
	svc, logger := newServiceTest(t, config.EnvProd, testRepoDir)

	_, err := svc.Index(context.Background())
	_, _ = svc.Index(context.Background())

	assert.ErrorIs(t, err, hperrors.ErrAppTemplatesUnavailable, "the signed source is used instead")
	assert.Len(t, logger.warnings, 1, "the warning is logged once, not per request")
}

func TestServiceRender(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	resp, err := svc.Render(context.Background(), &apptemplateservice.RenderReq{
		Name:   "demo",
		Params: map[string]any{"dataVolume": "vol-1"},
	})

	assert.NoError(t, err)
	assert.Equal(t, "demo:2.1.0", resp.Result.Image)
	assert.Equal(t, localSourceID, resp.Source)
	assert.Equal(t, "templates/demo.yaml", resp.Entry.File.Path)
	assert.NotEmpty(t, resp.Entry.File.SHA256)
}

func TestServiceRenderRefusesAnIncompatibleTemplate(t *testing.T) {
	dir := copyTestRepo(t)
	path := filepath.Join(dir, "templates", "demo.yaml")
	content, err := os.ReadFile(path)
	assert.NoError(t, err)
	assert.NoError(t, os.WriteFile(path, []byte(strings.Replace(string(content), "v000001", "v999999", 1)), 0o600))
	svc, _ := newServiceTest(t, config.EnvDev, dir)

	_, err = svc.Render(context.Background(), &apptemplateservice.RenderReq{
		Name: "demo", Params: map[string]any{"dataVolume": "vol-1"},
	})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateIncompatible)
}

func TestServiceTemplateNotFound(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	_, err := svc.Template(context.Background(), "nope")

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateNotFound)
}

func TestServiceIcon(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)
	index, err := svc.Index(context.Background())
	assert.NoError(t, err)
	sha := index.Index.FindTemplate("demo").Icon.SHA256

	icon, err := svc.Icon(context.Background(), &apptemplateservice.IconReq{
		Name: "demo", SHA256Prefix: sha[:apptemplateservice.IconHashLen], Ext: "svg",
	})

	assert.NoError(t, err)
	assert.Equal(t, "image/svg+xml", icon.ContentType)
	assert.Contains(t, string(icon.Content), "<svg")
}

// An icon URL names a template and a piece of its icon's hash. Anything that does
// not describe the icon the index lists now is not served: a URL is cached as
// immutable, so serving new bytes under an old hash would pin the wrong image.
func TestServiceIconRefusesWhatTheIndexDoesNotList(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)
	index, err := svc.Index(context.Background())
	assert.NoError(t, err)
	prefix := index.Index.FindTemplate("demo").Icon.SHA256[:apptemplateservice.IconHashLen]

	for name, req := range map[string]*apptemplateservice.IconReq{
		"unknown template":     {Name: "nope", SHA256Prefix: prefix, Ext: "svg"},
		"stale hash":           {Name: "demo", SHA256Prefix: "0000000000", Ext: "svg"},
		"another extension":    {Name: "demo", SHA256Prefix: prefix, Ext: "png"},
		"a shorter hash piece": {Name: "demo", SHA256Prefix: prefix[:4], Ext: "svg"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Icon(context.Background(), req)
			assert.ErrorIs(t, err, hperrors.ErrNotFound)
		})
	}
}

func TestServiceTemplateLoadsDependencies(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	resp, err := svc.Template(context.Background(), "demoweb")

	assert.NoError(t, err)
	assert.Len(t, resp.Dependencies, 1)
	assert.Equal(t, "db", resp.Dependencies[0].Dependency.Name)
	assert.Equal(t, "demo", resp.Dependencies[0].Template.Metadata.Name)
	assert.Equal(t, "templates/demo.yaml", resp.Dependencies[0].Entry.File.Path)
}

func TestServiceRenderRendersDependenciesFirst(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)

	resp, err := svc.Render(context.Background(), &apptemplateservice.RenderReq{
		Name:             "demoweb",
		AppName:          "blog",
		Params:           map[string]any{"dataVolume": "vol-web"},
		DependencyParams: map[string]map[string]any{"db": {"dataVolume": "vol-db"}},
	})

	assert.NoError(t, err)
	assert.Len(t, resp.Dependencies, 1)
	db := resp.Dependencies[0]
	assert.Equal(t, "db", db.Name)
	assert.Equal(t, "blog-db", db.AppName)
	assert.Equal(t, "vol-db", db.Render.Result.Doc.Deployment.Storage.Mounts["/data"].Source)
	assert.Equal(t, localRevision, db.Render.Revision, "one revision for every app of the request")

	envVars := resp.Result.Doc.Settings["envVars"].(map[string]any)["data"].([]any)
	assert.Equal(t, "${blog_db.HIVEPAAS_PASSWORD}", envVars[0].(map[string]any)["v"])
}

func TestServiceRenderRefusesWhatADependencyDoesNotAsk(t *testing.T) {
	svc, _ := newServiceTest(t, config.EnvDev, testRepoDir)
	cases := map[string]*apptemplateservice.RenderReq{
		"no app name": {
			Name: "demoweb", Params: map[string]any{"dataVolume": "v"},
			DependencyParams: map[string]map[string]any{"db": {"dataVolume": "v"}},
		},
		"unknown dependency": {
			Name: "demoweb", AppName: "blog", Params: map[string]any{"dataVolume": "v"},
			DependencyParams: map[string]map[string]any{"cache": {"dataVolume": "v"}},
		},
		"dependency parameters for a template without dependencies": {
			Name: "demo", AppName: "blog", Params: map[string]any{"dataVolume": "v"},
			DependencyParams: map[string]map[string]any{"db": {"dataVolume": "v"}},
		},
	}
	for name, req := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := svc.Render(context.Background(), req)
			assert.Error(t, err)
		})
	}
}
