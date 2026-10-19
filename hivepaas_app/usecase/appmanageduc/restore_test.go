package appmanageduc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/datakey"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

var (
	windowFrom = time.Date(2026, 10, 11, 2, 0, 0, 0, time.UTC)
	windowTo   = time.Date(2026, 10, 18, 11, 59, 0, 0, time.UTC)
)

func restorableStatus() *managedmodel.Status {
	return &managedmodel.Status{RestoreWindow: &managedmodel.TimeWindow{From: windowFrom, To: windowTo},
		Backups: []managedmodel.BackupRef{{Name: "base_000000010000000000000040", At: windowFrom},
			{Name: "base_000000010000000000000042", At: windowTo.Add(-10 * time.Hour)}}}
}

func restorablePostgres() *entity.ManagedDatabase {
	return &entity.ManagedDatabase{RestorePoints: []*entity.ManagedRestorePoint{
		{Name: "before-v42", Position: "0/3000060", At: windowTo.Add(-time.Hour)},
		{Name: "ancient", Position: "0/1000028", At: windowFrom.Add(-24 * time.Hour)},
	}}
}

func TestATargetInsideTheWindowIsResolved(t *testing.T) {
	inside := windowTo.Add(-2 * time.Hour)
	for name, tc := range map[string]struct {
		req     appmanageddto.ManagedRestoreTargetReq
		want    entity.ManagedRestoreTarget
		pointAt time.Time
	}{
		"latest": {appmanageddto.ManagedRestoreTargetReq{Kind: "latest"}, entity.ManagedRestoreTarget{Kind: "latest"},
			time.Time{}},
		"a time": {appmanageddto.ManagedRestoreTargetReq{Kind: "time", Time: &inside},
			entity.ManagedRestoreTarget{Kind: "time", Time: inside}, time.Time{}},
		"a point, by its LSN": {appmanageddto.ManagedRestoreTargetReq{Kind: "point", Point: "before-v42"},
			entity.ManagedRestoreTarget{Kind: "point", Point: "before-v42", Position: "0/3000060"}, windowTo.Add(-time.Hour)},
		"a backup": {appmanageddto.ManagedRestoreTargetReq{Kind: "backup", Backup: "base_000000010000000000000042"},
			entity.ManagedRestoreTarget{Kind: "backup", Backup: "base_000000010000000000000042"}, time.Time{}},
	} {
		pg := restorablePostgres()

		target, pointAt, err := resolveTarget(&tc.req, pg, restorableStatus())

		assert.NoError(t, err, name)
		assert.Equal(t, tc.want, target, name)
		assert.Equal(t, tc.pointAt, pointAt, name)
		assert.Equal(t, restorablePostgres(), pg, "%s: the source is only read", name)
	}
}

func TestATargetTheArchiveDoesNotReachIsRefused(t *testing.T) {
	before, after := windowFrom.Add(-time.Minute), windowTo.Add(time.Minute)
	for name, tc := range map[string]struct {
		req  appmanageddto.ManagedRestoreTargetReq
		want error
	}{
		"a time before the oldest backup": {appmanageddto.ManagedRestoreTargetReq{Kind: "time", Time: &before},
			hperrors.ErrAppManagedRestoreOutsideWindow},
		"a time after the last archived WAL": {appmanageddto.ManagedRestoreTargetReq{Kind: "time", Time: &after},
			hperrors.ErrAppManagedRestoreOutsideWindow},
		"a point gone": {appmanageddto.ManagedRestoreTargetReq{Kind: "point", Point: "never-made"},
			hperrors.ErrNotFound},
		"a point retention went past": {appmanageddto.ManagedRestoreTargetReq{Kind: "point", Point: "ancient"},
			hperrors.ErrAppManagedRestoreOutsideWindow},
		"a backup retention deleted": {appmanageddto.ManagedRestoreTargetReq{Kind: "backup",
			Backup: "base_000000010000000000000001"}, hperrors.ErrNotFound},
	} {
		_, _, err := resolveTarget(&tc.req, restorablePostgres(), restorableStatus())

		assert.ErrorIs(t, err, tc.want, name)
	}
}

func TestNothingIsRestoredBeforeTheFirstBackup(t *testing.T) {
	_, _, err := resolveTarget(&appmanageddto.ManagedRestoreTargetReq{Kind: "latest"}, restorablePostgres(),
		&managedmodel.Status{})

	assert.ErrorIs(t, err, hperrors.ErrAppManagedRestoreNoBackup)
}

// The new app is the source's template, version and parameters, with its own
// data volume: secrets decrypted, what the template no longer declares left
// out, what it has added left to its default.
func TestTheRestoredAppHasTheSourcesParameters(t *testing.T) {
	key, err := datakey.Generate()
	assert.NoError(t, err)
	datakey.SetActive(key)
	defs := []*templatemodel.Parameter{{Name: "dataVolume", Type: templatemodel.ParamTypeVolume},
		{Name: "password", Type: templatemodel.ParamTypeSecret}, {Name: "maxConnections", Type: "int"},
		{Name: "addedSince", Type: "string"}}
	source := map[string]*entity.AppTemplateParam{"dataVolume": {Value: "vol-src"},
		"password": {Secret: entity.NewEncryptedField("s3cret")}, "maxConnections": {Value: "200"},
		"removedSince": {Value: "x"}}

	params, err := restoreParams(defs, source, "vol-new")

	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"dataVolume": "vol-new", "password": "s3cret", "maxConnections": "200"}, params)

	_, err = restoreParams(append(defs, &templatemodel.Parameter{Name: "walVolume",
		Type: templatemodel.ParamTypeVolume}), source, "vol-new")
	assert.ErrorIs(t, err, hperrors.ErrAppTemplateIncompatible, "one volume is what a restore gives")
}

func TestTheRestoredAppArchivesWhereTheSourceDoesUnderItsOwnName(t *testing.T) {
	loaded := loadedPostgres(t)
	loaded.app = &entity.App{ID: "app-1", ProjectID: "p1", ProjectEnvID: "p1:prod", Name: "shop-db"}
	loaded.managed.Database.Storage.KeyAuth = entity.ObjectID{ID: "key-1"}
	loaded.managed.Database.RestoreTest.CheckQuery = "SELECT 1"
	binding := &entity.AppTemplateSettings{Template: "postgres-managed", Version: "17",
		ImageOverride: "docker.io/library/postgres:17.5-trixie"}
	defs := []*templatemodel.Parameter{{Name: "dataVolume", Type: templatemodel.ParamTypeVolume}}
	now := time.Date(2026, 10, 18, 14, 40, 7, 0, time.UTC)

	req, err := restoreCreateReq(loaded, binding, defs, &appmanageddto.RestoreManagedReq{
		DataVolume: &basedto.ObjectIDReq{ID: "vol-new"}}, now)

	assert.NoError(t, err)
	assert.Equal(t, "shop-db-restored-20261018-144007", req.Name)
	assert.Equal(t, "p1:prod", req.ProjectEnvID)
	assert.Equal(t, "postgres-managed", req.Template)
	assert.Equal(t, "17", req.Version)
	assert.Equal(t, "17.5-trixie", req.ImageTag, "the image the source was given")
	managed := req.Managed
	assert.Equal(t, "repo-1", managed.Storage.BackupRepo.ID)
	assert.Nil(t, managed.Storage.CloudStorage)
	assert.Equal(t, "key-1", managed.Storage.KeyAuth.ID)
	assert.Equal(t, 7, managed.RetainFull)
	assert.Equal(t, timeutil.Duration(time.Minute), managed.ArchiveTimeout)
	assert.Equal(t, "0 2 * * *", managed.BaseBackupSchedule)
	assert.Equal(t, "0 4 * * 0", managed.RestoreTestSchedule)
	assert.Equal(t, "SELECT 1", managed.RestoreTestQuery)
	assert.Equal(t, &apptemplatedto.ManagedDumpReq{}, managed.Dump, "the source is not dumped: neither is it")

	loaded.managed.Database.Dump = entity.ManagedDump{Repo: entity.ObjectID{ID: "repo-2"}, Schedule: "0 5 * * *"}
	req, err = restoreCreateReq(loaded, binding, defs, &appmanageddto.RestoreManagedReq{
		DataVolume: &basedto.ObjectIDReq{ID: "vol-new"}}, now)
	assert.NoError(t, err)
	assert.Equal(t, &apptemplatedto.ManagedDumpReq{Repo: &basedto.ObjectIDReq{ID: "repo-2"}, Schedule: "0 5 * * *"},
		req.Managed.Dump, "dumped where and when the source is")
}

func TestAnImagesTag(t *testing.T) {
	for image, tag := range map[string]string{
		"postgres:17.6-trixie":                   "17.6-trixie",
		"docker.io/library/postgres:17.6-trixie": "17.6-trixie",
		"registry.local:5000/pg":                 "",
		"registry.local:5000/pg:17@sha256:abc":   "17",
		"":                                       "",
	} {
		assert.Equal(t, tag, imageTag(image), image)
	}
}
