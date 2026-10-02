package imagebuildserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/registry"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryauthservice/registryauthserviceimpl"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingservice"
)

// withDataKey installs a data encryption key for the test, as the app has one.
func withDataKey(t *testing.T) {
	t.Helper()
	key, err := datakey.Generate()
	if err != nil {
		t.Fatal(err)
	}
	datakey.SetActive(key)
	t.Cleanup(func() { datakey.SetActive(nil) })
}

// sealed is a secret as the database holds it: only its ciphertext.
func sealed(t *testing.T, plain string) entity.EncryptedField {
	t.Helper()
	field := entity.NewEncryptedField(plain)
	encrypted, err := field.GetEncrypted()
	if err != nil {
		t.Fatal(err)
	}
	return entity.NewEncryptedField(encrypted)
}

func registrySetting(t *testing.T, id, address, username, password string) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Type: base.SettingTypeRegistryAuth, Status: base.SettingStatusActive}
	setting.MustSetData(&entity.RegistryAuth{Address: address, Username: username, Password: sealed(t, password)})
	return setting
}

type projectRegistries struct {
	repository.SettingRepo
	settings []*entity.Setting
}

func (r *projectRegistries) List(
	context.Context, database.IDB, *entity.ObjectScope, *basedto.Paging, ...bunex.SelectQueryOption,
) ([]*entity.Setting, *basedto.PagingMeta, error) {
	return r.settings, nil, nil
}

type refSettings struct {
	settingservice.Service
	settings map[string]*entity.Setting
}

func (r *refSettings) LoadRefObjectsByIDs(
	_ context.Context, _ database.IDB, refObjects **entity.RefObjects,
	_ *entity.ObjectScope, _ bool, _ *entity.RefObjectIDs,
) error {
	*refObjects = &entity.RefObjects{RefSettings: r.settings}
	return nil
}

type buildEnv struct {
	envvarservice.Service
	vars []*envvarservice.EnvVar
}

func (e *buildEnv) BuildEnvVarsInApp(
	context.Context, database.IDB, *envvarservice.BuildEnvVarsInAppReq,
) (*envvarservice.BuildEnvVarsInAppResp, error) {
	return &envvarservice.BuildEnvVarsInAppResp{EnvVars: e.vars}, nil
}

// Resolving a build's inputs opens every secret the build reads: the build
// variables, the project's registries, and the registry the image is pushed to.
func TestResolvingInputsOpensTheSecretsABuildReads(t *testing.T) {
	withDataKey(t)
	s := &service{
		registryAuthService: registryauthserviceimpl.New(nil, nil, nil, nil, nil, nil),
		settingRepo: &projectRegistries{settings: []*entity.Setting{
			registrySetting(t, "r1", "docker.io", "puller", "pull-pass"),
		}},
		settingService: &refSettings{settings: map[string]*entity.Setting{
			"r2": registrySetting(t, "r2", "registry.example.com", "hivepaas", "push-pass"),
		}},
		envVarService: &buildEnv{vars: []*envvarservice.EnvVar{{
			EnvVar:     &entity.EnvVar{Key: "NPM_TOKEN", Value: "tok-123"},
			RefSecrets: map[*entity.Secret]struct{}{{Key: "npm", Value: sealed(t, "tok-123")}: {}},
		}, {
			EnvVar: &entity.EnvVar{Key: "NODE_ENV", Value: "production"},
		}}},
	}

	inputs, err := s.ResolveBuildInputs(context.Background(), nil, &imagebuildservice.ImageBuildReq{
		App:            buildApp(),
		PushToRegistry: entity.ObjectID{ID: "r2"},
	})

	if !assert.NoError(t, err) || !assert.NotNil(t, inputs) {
		return
	}
	// A variable that uses a secret reaches the build as a secret, never as a
	// build argument: a build argument is written into the image's history.
	assert.Equal(t, map[string]string{"NPM_TOKEN": "tok-123"}, inputs.SecretEnvVars)
	assert.NotContains(t, inputs.EnvVars, "NPM_TOKEN")
	if assert.Contains(t, inputs.EnvVars, "NODE_ENV") {
		assert.Equal(t, "production", *inputs.EnvVars["NODE_ENV"])
	}
	assert.Equal(t, registry.AuthConfig{Username: "puller", Password: "pull-pass", ServerAddress: "docker.io"},
		inputs.RegistryAuths["docker.io"])
	assert.Equal(t, &registry.AuthConfig{
		Username: "hivepaas", Password: "push-pass", ServerAddress: "registry.example.com",
	}, inputs.PushRegistry)
	assert.ElementsMatch(t, []string{"tok-123", "pull-pass", "push-pass"}, inputs.Secrets)
}

// A registry to push to that is not there is an error before the build starts.
func TestResolvingInputsNeedsTheRegistryToPushTo(t *testing.T) {
	withDataKey(t)
	s := &service{
		registryAuthService: registryauthserviceimpl.New(nil, nil, nil, nil, nil, nil),
		settingRepo:         &projectRegistries{},
		settingService:      &refSettings{settings: map[string]*entity.Setting{}},
		envVarService:       &buildEnv{},
	}

	_, err := s.ResolveBuildInputs(context.Background(), nil, &imagebuildservice.ImageBuildReq{
		App:            buildApp(),
		PushToRegistry: entity.ObjectID{ID: "gone"},
	})

	assert.ErrorIs(t, err, hperrors.ErrMissing)
}

// A build given its inputs - one run by an agent, which has no key to open a
// secret with - reads no setting and opens nothing: it names the image and
// signs in to the registry with what it was given.
func TestABuildGivenItsInputsOpensNoSecret(t *testing.T) {
	datakey.SetActive(nil)
	given := &imagebuildservice.BuildInputs{
		EnvVars: map[string]*string{},
		PushRegistry: &registry.AuthConfig{
			Username: "hivepaas", Password: "push-pass", ServerAddress: "registry.example.com",
		},
	}
	// No repository and no service: touching a setting would panic.
	s := &service{}
	req := &imagebuildservice.ImageBuildReq{
		App: buildApp(), CommitHash: "9f3c1de0ab", PushToRegistry: entity.ObjectID{ID: "r2"}, Inputs: given,
	}

	inputs, err := s.buildInputs(context.Background(), nil, req)
	if !assert.NoError(t, err) {
		return
	}
	assert.Same(t, given, inputs)

	refs, err := buildImageReferences(req.App, req.CommitHash, nil, pushRegistry(inputs))
	assert.NoError(t, err)
	if assert.NotEmpty(t, refs) {
		assert.Contains(t, refs[0], "registry.example.com/hivepaas/")
	}

	header, err := pushAuthHeader(inputs)
	assert.NoError(t, err)
	assert.NotEmpty(t, header)
}
