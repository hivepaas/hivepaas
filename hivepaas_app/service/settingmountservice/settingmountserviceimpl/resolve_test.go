package settingmountserviceimpl

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
)

var testApp = &entity.App{ID: "app_1", GlobalKey: "shop_prod_api", ServiceID: "svc_1"}

var dataKeyOnce sync.Once

// useDataKey makes one data key active for the whole package: a source
// encrypted under one key cannot be read under the next.
func useDataKey(t *testing.T) {
	t.Helper()
	dataKeyOnce.Do(func() {
		key, err := datakey.Generate()
		assert.NoError(t, err)
		datakey.SetActive(key)
	})
}

func entry(t *testing.T, key string, status base.SettingStatus, data *entity.AppSettingMount) *entity.Setting {
	t.Helper()
	setting := &entity.Setting{ID: "entry_" + key, Type: base.SettingTypeAppSettingMount,
		Scope: base.ObjectScopeApp, ObjectID: testApp.ID, Name: key, Status: status}
	assert.NoError(t, setting.SetData(data))
	return setting
}

func certSource(t *testing.T, id, cert, key string) *entity.Setting {
	t.Helper()
	useDataKey(t)
	setting := &entity.Setting{ID: id, Type: base.SettingTypeSSLCert, Status: base.SettingStatusActive}
	assert.NoError(t, setting.SetData(&entity.SSLCert{Certificate: cert,
		PrivateKey: entity.NewEncryptedField(key)}))
	return setting
}

func certFiles(source string) *entity.AppSettingMount {
	return &entity.AppSettingMount{Source: entity.ObjectID{ID: source}, Files: []*entity.AppSettingMountFile{
		{Part: "certificate", Path: "/etc/app/tls/cert.pem"},
		{Part: "privateKey", Path: "/etc/app/tls/key.pem", UID: "1000", Mode: 0o400},
		{Part: "caCertificate", Path: "/etc/app/tls/ca.pem"},
	}}
}

// fixture is a service whose seams answer from memory: entries of the app, and
// the sources its scope sees.
func fixture(t *testing.T, entries []*entity.Setting, sources ...*entity.Setting) *service {
	t.Helper()
	useDataKey(t)
	svc := &service{rotationKey: func() []byte { return []byte("k") }, removalRetryDelay: time.Millisecond}
	svc.loadEntries = func(context.Context, database.IDB, *entity.App) ([]*entity.Setting, error) {
		return entries, nil
	}
	svc.loadSources = func(_ context.Context, _ database.IDB, _ *entity.App, ids []string) ([]*entity.Setting, error) {
		var out []*entity.Setting
		for _, source := range sources {
			for _, id := range ids {
				if source.ID == id {
					out = append(out, source)
				}
			}
		}
		return out, nil
	}
	return svc
}

func paths(files []*settingmountservice.File) []string {
	var out []string
	for _, file := range files {
		out = append(out, file.Path)
	}
	return out
}

func TestResolveRendersAnEntrysFiles(t *testing.T) {
	svc := fixture(t, []*entity.Setting{entry(t, "tls-cert", base.SettingStatusActive, certFiles("cert_1"))},
		certSource(t, "cert_1", "CERT", "KEY"))

	files, err := svc.Resolve(context.Background(), nil, testApp)

	assert.NoError(t, err)
	if !assert.Len(t, files, 2, "no CA certificate, so no ca.pem") {
		return
	}
	assert.Equal(t, &settingmountservice.File{Entry: "tls-cert", Part: "certificate", Path: "/etc/app/tls/cert.pem",
		UID: "0", GID: "0", Mode: fileutil.FileMode(0o444), Data: []byte("CERT"), Rotation: files[0].Rotation},
		files[0])
	assert.Equal(t, "1000", files[1].UID)
	assert.Equal(t, fileutil.FileMode(0o400), files[1].Mode)
	assert.True(t, files[1].Secret)
	assert.Equal(t, []byte("KEY"), files[1].Data)
	assert.Len(t, files[0].Rotation, 64)
}

func TestResolveLeavesOutWhatCannotBeUsed(t *testing.T) {
	for name, tc := range map[string]struct {
		entry   *entity.Setting
		sources []*entity.Setting
	}{
		"a disabled entry": {entry(t, "a", base.SettingStatusDisabled, certFiles("cert_1")),
			[]*entity.Setting{certSource(t, "cert_1", "CERT", "KEY")}},
		"a source the scope does not return mounts nothing": {entry(t, "a", base.SettingStatusActive,
			certFiles("cert_other")), []*entity.Setting{certSource(t, "cert_1", "CERT", "KEY")}},
		"a certificate not obtained yet": {entry(t, "a", base.SettingStatusActive, certFiles("cert_1")),
			[]*entity.Setting{certSource(t, "cert_1", "", "")}},
		"a key no entry may have": {entry(t, "Cert", base.SettingStatusActive, certFiles("cert_1")),
			[]*entity.Setting{certSource(t, "cert_1", "CERT", "KEY")}},
	} {
		svc := fixture(t, []*entity.Setting{tc.entry}, tc.sources...)
		files, err := svc.Resolve(context.Background(), nil, testApp)
		assert.NoError(t, err, name)
		assert.Empty(t, files, name)
	}
	disabled := certSource(t, "cert_1", "CERT", "KEY")
	disabled.Status = base.SettingStatusDisabled
	svc := fixture(t, []*entity.Setting{entry(t, "a", base.SettingStatusActive, certFiles("cert_1"))}, disabled)
	files, err := svc.Resolve(context.Background(), nil, testApp)
	assert.NoError(t, err)
	assert.Empty(t, files, "a disabled source, should the scope return it")
}

func TestResolveSkipsAFileThatIsWrongInItself(t *testing.T) {
	svc := fixture(t, []*entity.Setting{entry(t, "a", base.SettingStatusActive, &entity.AppSettingMount{
		Source: entity.ObjectID{ID: "cert_1"}, Files: []*entity.AppSettingMountFile{
			{Part: "certificate", Path: "relative/cert.pem"},
			{Part: "certificate", Path: "/etc/app/../cert.pem"},
			{Part: "htpasswd", Path: "/etc/htpasswd"},
			{Part: "privateKey", Path: "/etc/key.pem"},
		}})}, certSource(t, "cert_1", "CERT", "KEY"))

	files, err := svc.Resolve(context.Background(), nil, testApp)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/etc/key.pem"}, paths(files))
}

// Entries are taken in key order and the first claim to a path wins: which file
// lands there never depends on the order rows come back in.
func TestTwoEntriesClaimingAPathKeepTheFirstByKey(t *testing.T) {
	second := entry(t, "b", base.SettingStatusActive, certFiles("cert_2"))
	first := entry(t, "a", base.SettingStatusActive, certFiles("cert_1"))
	svc := fixture(t, []*entity.Setting{second, first},
		certSource(t, "cert_1", "ONE", "K1"), certSource(t, "cert_2", "TWO", "K2"))

	files, err := svc.Resolve(context.Background(), nil, testApp)

	assert.NoError(t, err)
	assert.Equal(t, []string{"/etc/app/tls/cert.pem", "/etc/app/tls/key.pem"}, paths(files))
	assert.Equal(t, []byte("ONE"), files[0].Data)
}

// Why an entry is not in the container is what the screen says.
func TestEntryStatesSayWhyNothingIsMounted(t *testing.T) {
	mounted := entry(t, "a", base.SettingStatusActive, certFiles("cert_1"))
	disabled := entry(t, "b", base.SettingStatusDisabled, certFiles("cert_1"))
	missing := entry(t, "c", base.SettingStatusActive, certFiles("cert_gone"))
	incomplete := entry(t, "d", base.SettingStatusActive, certFiles("cert_2"))
	shadowed := entry(t, "e", base.SettingStatusActive, certFiles("cert_1"))
	svc := fixture(t, []*entity.Setting{mounted, disabled, missing, incomplete, shadowed},
		certSource(t, "cert_1", "CERT", "KEY"), certSource(t, "cert_2", "", ""))

	states, err := svc.EntryStates(context.Background(), nil, testApp)

	assert.NoError(t, err)
	assert.Equal(t, &settingmountservice.EntryState{Mounted: []string{"/etc/app/tls/cert.pem", "/etc/app/tls/key.pem"}},
		states[mounted.ID])
	assert.Equal(t, settingmountservice.ReasonEntryDisabled, states[disabled.ID].Reason)
	assert.Equal(t, settingmountservice.ReasonSourceUnavailable, states[missing.ID].Reason)
	assert.Equal(t, settingmountservice.ReasonSourceIncomplete, states[incomplete.ID].Reason)
	assert.Equal(t, settingmountservice.ReasonPathsTaken, states[shadowed.ID].Reason)
}

// A preview's entries are its parent's inheritable ones; the files carry the
// preview's own name.
func TestAPreviewResolvesItsParentsInheritableEntries(t *testing.T) {
	inherited := entry(t, "cert", base.SettingStatusActive, certFiles("cert_1"))
	inherited.Inheritable = true
	svc := fixture(t, []*entity.Setting{inherited}, certSource(t, "cert_1", "CERT", "KEY"))
	var asked *entity.App
	loadAll := svc.loadEntries
	svc.loadEntries = func(ctx context.Context, db database.IDB, app *entity.App) ([]*entity.Setting, error) {
		asked = app
		return loadAll(ctx, db, app)
	}
	preview := &entity.App{ID: "app_preview", ParentID: testApp.ID, GlobalKey: "shop_prod_api-pr-7"}

	files, err := svc.Resolve(context.Background(), nil, preview)

	assert.NoError(t, err)
	assert.Equal(t, preview, asked, "the preview is who asks: its parent's entries are the repository's to add")
	assert.Len(t, files, 2)
}
