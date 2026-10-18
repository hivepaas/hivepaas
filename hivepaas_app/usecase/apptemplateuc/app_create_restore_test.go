package apptemplateuc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

type fakeTaskRepo struct {
	repository.TaskRepo
	inserted []*entity.Task
}

func (f *fakeTaskRepo) Insert(_ context.Context, _ database.IDB, task *entity.Task,
	_ ...bunex.InsertQueryOption) error {
	f.inserted = append(f.inserted, task)
	return nil
}

func restoreSource() *managedmodel.RestoreSource {
	return &managedmodel.RestoreSource{SourceAppID: "app-src", EncryptionKey: "source-key",
		Archive: &managedmodel.Archive{S3: &managedmodel.S3Archive{Bucket: "b", Prefix: "walg/shop-db"}},
		Target:  entity.ManagedRestoreTarget{Kind: entity.ManagedRestoreToLatest}}
}

// Restoring is never gated: a template needing a feature the license lacks is
// restored all the same.
func TestARestoreAsksTheLicenseNothing(t *testing.T) {
	uc, fakes := newCreateTest(t)
	fakes.templates.resp.Template.Metadata.Requires.Features = []string{"managed-databases"}
	fakes.licenses.featuresErr = hperrors.Wrap(hperrors.ErrPlanFeatureRequired)
	req := testCreateReq()
	apps := planApps(req, fakes.templates.resp)

	err := uc.checkRequest(context.Background(), testAuth(), req, fakes.templates.resp, apps,
		&managedRestore{source: restoreSource()})

	assert.NoError(t, err)
	assert.Empty(t, fakes.licenses.features)
	assert.ErrorIs(t, uc.checkRequest(context.Background(), testAuth(), req, fakes.templates.resp, apps, nil),
		hperrors.ErrPlanFeatureRequired, "a creation still asks")
}

func TestTheEngineIsGivenWhatTheAppRecoversFrom(t *testing.T) {
	uc, fakes := newManagedTest(t)
	req := managedReq(apptemplatedto.ManagedStorageReq{BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}})
	source := restoreSource()

	planned, err := uc.planManaged(context.Background(), req, planApps(req, fakes.templates.resp),
		&managedRestore{source: source})

	assert.NoError(t, err)
	assert.Same(t, source, planned.prepared.Restore)
	if assert.NotNil(t, fakes.engine.prepared) {
		assert.Same(t, source, fakes.engine.prepared.Restore, "before the document is prepared")
	}
}

func TestOnlyAManagedTemplateIsRestored(t *testing.T) {
	uc, fakes := newCreateTest(t)
	req := testCreateReq()

	_, err := uc.planManaged(context.Background(), req, planApps(req, fakes.templates.resp),
		&managedRestore{source: restoreSource()})

	assert.ErrorIs(t, err, hperrors.ErrAppTemplateParamInvalid)
}

// The task that sees the restore through is made with the app, in the same
// transaction, on the app restored into.
func TestARestoreMakesItsTaskWithTheApp(t *testing.T) {
	uc, fakes := newManagedTest(t)
	tasks := &fakeTaskRepo{}
	uc.taskRepo = tasks
	req := managedReq(apptemplatedto.ManagedStorageReq{BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}})
	apps := planApps(req, fakes.templates.resp)
	restore := &managedRestore{source: restoreSource()}
	planned, err := uc.planManaged(context.Background(), req, apps, restore)
	assert.NoError(t, err)
	planned.attach(apps)

	provisioned, err := uc.provisionAll(context.Background(), database.Tx{}, testAuth(), req, apps, restore)

	assert.NoError(t, err)
	app := provisioned.Apps[0].App
	if assert.Len(t, tasks.inserted, 1) && assert.Same(t, tasks.inserted[0], restore.task) {
		task := restore.task
		assert.Equal(t, base.TaskTypeManagedRestore, task.Type)
		assert.Equal(t, app.ID, task.TargetID)
		assert.Equal(t, app.ID, task.ObjectID)
		args, _ := task.ArgsAsManagedRestore()
		assert.Equal(t, &entity.TaskManagedRestoreArgs{ProjectID: "p1", AppID: app.ID, SourceAppID: "app-src"}, args)
		state, _ := task.OutputAsManagedRestore()
		assert.Equal(t, entity.ManagedRestoreStepRecover, state.Step)
		assert.Positive(t, task.Config.MaxRetry, "each step can run again")
	}
}

func TestAnOrdinaryCreationMakesNoRestoreTask(t *testing.T) {
	uc, fakes := newManagedTest(t)
	tasks := &fakeTaskRepo{}
	uc.taskRepo = tasks
	req := managedReq(apptemplatedto.ManagedStorageReq{BackupRepo: &basedto.ObjectIDReq{ID: "repo-s3"}})
	apps := planApps(req, fakes.templates.resp)
	planned, err := uc.planManaged(context.Background(), req, apps, nil)
	assert.NoError(t, err)
	planned.attach(apps)

	_, err = uc.provisionAll(context.Background(), database.Tx{}, testAuth(), req, apps, nil)

	assert.NoError(t, err)
	assert.Empty(t, tasks.inserted)
}

func TestARestoreNeedsItsArchiveAndSource(t *testing.T) {
	uc, _ := newCreateTest(t)

	_, _, err := uc.RestoreManagedApp(context.Background(), testAuth(), testCreateReq(), restoreSource())

	assert.ErrorIs(t, err, hperrors.ErrInternal)
}
