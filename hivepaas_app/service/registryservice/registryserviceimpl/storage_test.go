package registryserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// memSettingRepo holds settings by id. List answers with all of them: the
// filtering the real query does is repeated by the code under test, which is
// what is being checked.
type memSettingRepo struct {
	repository.SettingRepo
	settings map[string]*entity.Setting
}

func (r *memSettingRepo) GetByID(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ base.SettingType,
	id string, _ bool, _ ...bunex.SelectQueryOption) (*entity.Setting, error) {
	if setting, ok := r.settings[id]; ok {
		return setting, nil
	}
	return nil, hperrors.ErrNotFound
}

func (r *memSettingRepo) List(_ context.Context, _ database.IDB, _ *entity.ObjectScope, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.Setting, *basedto.PagingMeta, error) {
	out := make([]*entity.Setting, 0, len(r.settings))
	for _, setting := range r.settings {
		out = append(out, setting)
	}
	return out, nil, nil
}

func (r *memSettingRepo) Upsert(_ context.Context, _ database.IDB, setting *entity.Setting,
	_, _ []string, _ ...bunex.InsertQueryOption) error {
	r.settings[setting.ID] = setting
	return nil
}

func registryAuths(r *memSettingRepo) []*entity.Setting {
	var out []*entity.Setting
	for _, setting := range r.settings {
		if setting.Type == base.SettingTypeRegistryAuth {
			out = append(out, setting)
		}
	}
	return out
}

func mustRegistryAuth(t *testing.T, setting *entity.Setting) *entity.RegistryAuth {
	t.Helper()
	auth, err := setting.AsRegistryAuth()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return auth
}

// A credential made by an operator, by hand: no marker.
func operatorCredential(t *testing.T, id, address, password string) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: id, Scope: base.ObjectScopeGlobal, Type: base.SettingTypeRegistryAuth,
		Kind: address, Name: "Mine", Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(&entity.RegistryAuth{
		Username: registryUsername, Password: entity.NewEncryptedField(password), Address: address,
	}))
	return setting
}

// The reported bug: the registry is switched off while an app still names its
// credential, then switched on again - at another address. The credential it
// had is taken up again, with its password, so the app keeps pulling; no second
// one is made beside it.
func TestSwitchingBackOnTakesUpTheCredentialItHad(t *testing.T) {
	useDataKey(t)
	first, err := newRegistryAuthSetting("cred-1", "registry.old.example", "the-password", time.Now())
	assert.NoError(t, err)
	repo := &memSettingRepo{settings: map[string]*entity.Setting{"cred-1": first}}
	s := &service{settingRepo: repo}

	setting, password, err := s.ensureCredential(context.Background(), nil,
		&entity.RegistrySettings{Domain: "registry.new.example", RegistryAuthID: "cred-1"})

	assert.NoError(t, err)
	assert.Equal(t, "cred-1", setting.ID)
	assert.Equal(t, "the-password", password)
	assert.Equal(t, "registry.new.example", mustRegistryAuth(t, setting).Address)
	assert.Equal(t, "registry.new.example", setting.Kind)
	assert.Len(t, registryAuths(repo), 1, "no second credential")
}

// A registry that forgot the id - switched off before the id was kept - finds
// its credential by the marker, whatever it is called and addressed at now.
func TestTheCredentialIsFoundByItsMarkerWhenTheIdIsGone(t *testing.T) {
	useDataKey(t)
	own, err := newRegistryAuthSetting("cred-1", "registry.old.example", "the-password", time.Now())
	assert.NoError(t, err)
	own.Name = "Renamed by somebody"
	repo := &memSettingRepo{settings: map[string]*entity.Setting{"cred-1": own}}
	s := &service{settingRepo: repo}

	setting, password, err := s.ensureCredential(context.Background(), nil,
		&entity.RegistrySettings{Domain: "registry.new.example"})

	assert.NoError(t, err)
	assert.Equal(t, "cred-1", setting.ID)
	assert.Equal(t, "the-password", password)
	assert.Len(t, registryAuths(repo), 1)
}

// An operator's credential for the same account at the same address is the
// operator's: the registry makes its own rather than taking it over.
func TestAnOperatorsCredentialIsNotTakenOver(t *testing.T) {
	useDataKey(t)
	repo := &memSettingRepo{settings: map[string]*entity.Setting{
		"mine": operatorCredential(t, "mine", "registry.example", "operator-password"),
	}}
	s := &service{settingRepo: repo}

	setting, password, err := s.ensureCredential(context.Background(), nil,
		&entity.RegistrySettings{Domain: "registry.example"})

	assert.NoError(t, err)
	assert.NotEqual(t, "mine", setting.ID)
	assert.NotEqual(t, "operator-password", password)
	assert.Equal(t, entity.RegistryAuthManagedBySystemRegistry, mustRegistryAuth(t, setting).ManagedBy)
	assert.Len(t, registryAuths(repo), 2)
}

// A credential the registry made before the marker existed is marked the next
// time it is used, so it can be found again once its id is forgotten.
func TestACredentialMadeBeforeTheMarkerIsMarked(t *testing.T) {
	useDataKey(t)
	legacy := operatorCredential(t, "cred-1", "registry.example", "the-password")
	repo := &memSettingRepo{settings: map[string]*entity.Setting{"cred-1": legacy}}
	s := &service{settingRepo: repo}

	setting, _, err := s.ensureCredential(context.Background(), nil,
		&entity.RegistrySettings{Domain: "registry.example", RegistryAuthID: "cred-1"})

	assert.NoError(t, err)
	assert.Equal(t, entity.RegistryAuthManagedBySystemRegistry, mustRegistryAuth(t, setting).ManagedBy)
}

// Among several marked - duplicates left by switch-ons before this was fixed -
// the one changed last is taken.
func TestPickOwnCredentialTakesTheLatestMarked(t *testing.T) {
	useDataKey(t)
	older, _ := newRegistryAuthSetting("older", "r.example", "a", time.Now().Add(-time.Hour))
	newer, _ := newRegistryAuthSetting("newer", "r.example", "b", time.Now())
	mine := operatorCredential(t, "mine", "r.example", "c")
	mine.UpdatedAt = time.Now().Add(time.Hour)

	assert.Equal(t, "newer", pickOwnCredential([]*entity.Setting{older, mine, newer}).ID)
	assert.Nil(t, pickOwnCredential([]*entity.Setting{mine}))
}

// quietLogger takes the one line teardown logs.
type quietLogger struct{ logging.Logger }

func (quietLogger) Info(string, ...any) {}

// fakeSystemApps has the registry's app running, or none, and removes it when asked.
type fakeSystemApps struct {
	systemappservice.Service
	app     *entity.App
	removed bool
}

func (f *fakeSystemApps) LoadApp(context.Context, database.IDB, string) (*entity.App, error) {
	return f.app, nil
}

func (f *fakeSystemApps) Remove(context.Context, database.IDB, *entity.App, bool) error {
	f.removed = true
	return nil
}

func registrySetting(t *testing.T, cfg *entity.RegistrySettings) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "registry", Scope: base.ObjectScopeGlobal, Type: base.SettingTypeRegistry}
	assert.NoError(t, setting.SetData(cfg))
	return setting
}

// Switching the registry off forgets its app but not its credential: the next
// switch-on finds the credential by that id.
func TestTeardownKeepsTheCredentialsId(t *testing.T) {
	for _, running := range []bool{true, false} {
		apps := &fakeSystemApps{}
		if running {
			apps.app = &entity.App{ID: "app-1"}
		}
		repo := &memSettingRepo{settings: map[string]*entity.Setting{}}
		s := &service{settingRepo: repo, systemAppService: apps, logger: quietLogger{}}
		cfg := &entity.RegistrySettings{AppID: "app-1", RegistryAuthID: "cred-1"}
		setting := registrySetting(t, cfg)

		resp, err := s.teardown(context.Background(), nil, setting, cfg,
			&registryservice.SettingApplyReq{RemoveApp: true})

		assert.NoError(t, err)
		stored, err := setting.AsRegistrySettings()
		assert.NoError(t, err)
		assert.Empty(t, stored.AppID)
		assert.Equal(t, "cred-1", stored.RegistryAuthID)
		assert.Equal(t, running, apps.removed)
		if running {
			assert.Equal(t, "cred-1", resp.RemovedCredentialID, "offered for deletion if nothing names it")
		}
	}
}
