package appuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appdeploymentservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

// fakeSettingService loads the settings a function's source refers to, and
// keeps where it was asked to look.
type fakeSettingService struct {
	settingservice.Service
	loaded map[string]*entity.Setting
	scope  *entity.ObjectScope
	refIDs *entity.RefObjectIDs
}

func (f *fakeSettingService) LoadRefObjectsByIDs(
	_ context.Context, _ database.IDB, refObjects **entity.RefObjects,
	scope *entity.ObjectScope, _ bool, refIDs *entity.RefObjectIDs,
) error {
	f.scope, f.refIDs = scope, refIDs
	if *refObjects == nil {
		*refObjects = entity.NewRefObjects()
	}
	for id, setting := range f.loaded {
		(*refObjects).RefSettings[id] = setting
	}
	return nil
}

// fakeDeploymentService keeps the build source it was asked to check.
type fakeDeploymentService struct {
	appdeploymentservice.Service
	checked *appdeploymentservice.CheckBuildSourceReq
}

func (f *fakeDeploymentService) CheckBuildSource(
	_ context.Context, req *appdeploymentservice.CheckBuildSourceReq,
) error {
	f.checked = req
	return nil
}

// A function's source is checked in its environment before anything is
// created: the settings it refers to, and what its build will need.
func TestAFunctionsSourceIsCheckedInItsEnvironmentFirst(t *testing.T) {
	credentials := &entity.Setting{ID: "cred-1", Type: base.SettingTypeAccessToken}
	settings := &fakeSettingService{loaded: map[string]*entity.Setting{"cred-1": credentials}}
	deployments := &fakeDeploymentService{}
	uc := &UC{settingService: settings, appDeploymentService: deployments}
	source := &entity.DeploymentFunctionSource{
		Runtime: base.FunctionRuntimeGo127,
		Code: entity.FunctionCode{Dir: "fns/hello", Repo: &entity.FunctionRepoCode{
			RepoType: base.RepoTypeGit, RepoURL: "https://github.com/acme/fns.git", RepoRef: "refs/heads/main",
			Credentials: entity.RepoCredentials{ID: "cred-1"},
		}},
		PushToRegistry: entity.ObjectID{ID: "registry-1"},
	}

	err := uc.checkFunctionSource(context.Background(), "project-1", "env-1", source)

	assert.NoError(t, err)
	assert.Equal(t, base.ObjectScopeProjectEnv, settings.scope.ScopeType)
	assert.Equal(t, "project-1", settings.scope.ProjectID)
	assert.Equal(t, "env-1", settings.scope.ProjectEnvID)
	assert.ElementsMatch(t, []string{"cred-1", "registry-1"}, settings.refIDs.RefSettingIDs)
	assert.Equal(t, "https://github.com/acme/fns.git", deployments.checked.RepoSource.RepoURL)
	assert.Equal(t, "registry-1", deployments.checked.PushToRegistry.ID)
	assert.Same(t, credentials, deployments.checked.RefObjects.RefSettings["cred-1"])
}

// A function is created with its kind, its source to deploy, and its routing
// at the port its runtime listens on.
func TestAFunctionIsCreatedWithItsKindSourceAndPort(t *testing.T) {
	app := &entity.App{ID: "app-1"}
	source := &entity.DeploymentFunctionSource{Runtime: base.FunctionRuntimeNode24}
	timeNow := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)

	settings := functionSettings(app, source, "", timeNow)

	byType := map[base.SettingType]*entity.Setting{}
	for _, setting := range settings {
		assert.Equal(t, "app-1", setting.ObjectID)
		assert.Equal(t, base.ObjectScopeApp, setting.Scope)
		assert.Equal(t, base.SettingStatusActive, setting.Status)
		assert.Equal(t, timeNow, setting.CreatedAt)
		assert.NotEmpty(t, setting.ID)
		byType[setting.Type] = setting
	}
	assert.Len(t, byType, 3)
	assert.True(t, entity.IsFunctionKind(byType[base.SettingTypeAppKind]))
	deployment := byType[base.SettingTypeAppDeployment].MustAsAppDeploymentSettings()
	assert.Equal(t, base.DeploymentMethodFunction, deployment.ActiveMethod)
	assert.Equal(t, base.FunctionRuntimeNode24, deployment.FunctionSource.Runtime)
	routing := byType[base.SettingTypeAppRouting].MustAsAppRoutingSettings()
	assert.Equal(t, base.FunctionPort, routing.Port)
	assert.False(t, routing.ExposePublicly, "a function is not public unless asked")
	assert.Empty(t, routing.Domains)
	// What a clone copies: its deployment and routing settings, as an app's.
	assert.True(t, byType[base.SettingTypeAppDeployment].Inheritable)
	assert.True(t, byType[base.SettingTypeAppRouting].Inheritable)
}

// A function created at a domain is routed there from its first deployment:
// public, over HTTPS, at its runtime's port.
func TestAFunctionAskedForADomainIsRoutedThere(t *testing.T) {
	app := &entity.App{ID: "app-1"}
	source := &entity.DeploymentFunctionSource{Runtime: base.FunctionRuntimeNode24}

	settings := functionSettings(app, source, "hello.example.com", time.Now())

	var routing *entity.AppRoutingSettings
	for _, setting := range settings {
		if setting.Type == base.SettingTypeAppRouting {
			routing = setting.MustAsAppRoutingSettings()
		}
	}
	if !assert.NotNil(t, routing) {
		return
	}
	assert.True(t, routing.ExposePublicly)
	assert.Equal(t, []*entity.AppDomain{{
		Enabled:       true,
		Domain:        "hello.example.com",
		Protocol:      base.NetworkProtocolHTTP,
		ContainerPort: base.FunctionPort,
		ForceHttps:    true,
	}}, routing.Domains)
	assert.Equal(t, []string{"hello.example.com"}, routing.GetActiveDomainNames())
}

// fakeDomainService says whether a project may use a domain and whether it is
// free, and keeps what it was asked.
type fakeDomainService struct {
	domainservice.Service
	notAllowed, taken error
	verified, checked []string
	projectID         string
}

func (f *fakeDomainService) VerifyProjectDomains(
	_ context.Context, _ database.IDB, projectID string, domains []string,
) error {
	f.projectID, f.verified = projectID, domains
	return f.notAllowed
}

func (f *fakeDomainService) VerifyDomainsAvailable(
	_ context.Context, _ database.IDB, domains []string, _ []string,
) error {
	f.checked = domains
	return f.taken
}

// A function's domain is checked before anything is created, as a template's
// domains are: the project may use it, and no other app holds it.
func TestAFunctionsDomainIsCheckedInItsProjectFirst(t *testing.T) {
	domains := &fakeDomainService{}
	uc := &UC{domainService: domains}

	assert.NoError(t, uc.checkFunctionDomain(context.Background(), "project-1", "hello.example.com"))
	assert.Equal(t, "project-1", domains.projectID)
	assert.Equal(t, []string{"hello.example.com"}, domains.verified)
	assert.Equal(t, []string{"hello.example.com"}, domains.checked)

	domains = &fakeDomainService{taken: hperrors.Wrap(hperrors.ErrDomainInUse)}
	uc = &UC{domainService: domains}
	assert.ErrorIs(t, uc.checkFunctionDomain(context.Background(), "project-1", "hello.example.com"),
		hperrors.ErrDomainInUse)

	domains = &fakeDomainService{}
	uc = &UC{domainService: domains}
	assert.NoError(t, uc.checkFunctionDomain(context.Background(), "project-1", ""))
	assert.Nil(t, domains.verified, "no domain, nothing to check")
}
