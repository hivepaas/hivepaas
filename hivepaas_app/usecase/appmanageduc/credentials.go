package appmanageduc

import (
	"context"
	"maps"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/managedmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// RefreshManagedCredentials gives a managed app the credentials its archive's
// storage has now. They were copied into the app's secrets when it was made, so
// a key rotated on the storage, or the narrower key, reaches the database only
// this way - and its verifier app's copies, under the verifier's own key. The
// secrets that changed are mounted again, which restarts the database's task;
// when none did, nothing happens.
func (uc *UC) RefreshManagedCredentials(
	ctx context.Context,
	auth *basedto.Auth,
	req *appmanageddto.RefreshManagedCredentialsReq,
) (*appmanageddto.RefreshManagedCredentialsResp, error) {
	var refresh *entity.Task
	resp := &appmanageddto.RefreshManagedCredentialsResp{Data: &appmanageddto.RefreshManagedCredentialsDataResp{}}
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		loaded, err := uc.loadManaged(ctx, db, req.ProjectID, req.AppID, bunex.SelectFor("UPDATE OF app"))
		if err != nil {
			return err
		}
		if loaded.managed.Postgres == nil {
			return hperrors.NewNotFound("Managed archive")
		}
		var changed []*entity.Setting
		for _, reader := range archiveReaders(loaded.app, loaded.managed.Postgres) {
			refreshed, refreshErr := uc.refreshArchiveSecrets(ctx, db, loaded, reader)
			if refreshErr != nil {
				return refreshErr
			}
			changed = append(changed, refreshed...)
		}
		if len(changed) == 0 {
			return nil
		}
		now := timeutil.NowUTC()
		for _, setting := range changed {
			setting.UpdatedAt = now
			if err = uc.settingRepo.Update(ctx, db, setting,
				bunex.UpdateColumns("data", "update_ver", "updated_at")); err != nil {
				return hperrors.Wrap(err)
			}
		}
		if refresh, err = uc.settingMountService.RecordRefresh(ctx, db, changed...); err != nil {
			return hperrors.Wrap(err)
		}
		resp.Data.Changed = true
		return hperrors.Wrap(auditservice.RecordAllowed(ctx, uc.auditService, db, &auditservice.Entry{
			Type:     base.AuditLogTypeAppUpdate,
			Scope:    base.ObjectScopeApp,
			ObjectID: loaded.app.ID,
			Source:   base.AuditLogSourceAPIAction,
			Section:  "managed-credentials",
			Auth:     auth,
			ResType:  base.ResourceTypeApp,
			ResID:    loaded.app.ID,
			ResName:  loaded.app.Name,
			Detail:   auditdetail.New().Set("changedSecrets", len(changed)).String(),
		}))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if refresh != nil {
		// Once the secrets are committed: the refresh reads them.
		if err = uc.taskQueue.ScheduleTask(ctx, refresh); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return resp, nil
}

// archiveReader is an app holding copies of an archive's credentials, and the
// storage, with the key, they are copies of.
type archiveReader struct {
	appID   string
	storage entity.ManagedStorage
}

// archiveReaders are the database, with its storage, and its verifier app when
// it has one, with the verifier's narrower key when it was given one.
func archiveReaders(app *entity.App, pg *entity.ManagedPostgres) []archiveReader {
	readers := []archiveReader{{appID: app.ID, storage: pg.Storage}}
	if verifier := pg.RestoreTest.VerifierApp.ID; verifier != "" {
		storage := pg.Storage
		if pg.RestoreTest.KeyAuth.ID != "" {
			storage.KeyAuth = pg.RestoreTest.KeyAuth
		}
		readers = append(readers, archiveReader{appID: verifier, storage: storage})
	}
	return readers
}

// refreshArchiveSecrets gives an app's copies of the archive's credentials the
// values the storage has now, and returns those that changed.
func (uc *UC) refreshArchiveSecrets(
	ctx context.Context,
	db database.IDB,
	loaded *loadedManaged,
	reader archiveReader,
) ([]*entity.Setting, error) {
	s3, err := uc.managedService.ResolveS3(ctx, db, appScope(loaded.app), reader.storage)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	values := loaded.engine.ArchiveSecrets(&managedmodel.Archive{S3: s3})
	secrets, _, err := uc.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeSecret),
		bunex.SelectWhere("setting.object_id = ?", reader.appID),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectWhereIn("setting.name IN (?)", slices.Sorted(maps.Keys(values))...),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return refreshedSecrets(secrets, values)
}

// refreshedSecrets gives each secret its new value, and returns those whose
// value changed.
func refreshedSecrets(secrets []*entity.Setting, values map[string]string) ([]*entity.Setting, error) {
	var changed []*entity.Setting
	for _, setting := range secrets {
		value, found := values[setting.Name]
		if !found {
			continue
		}
		secret, err := setting.AsSecret()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		current, err := secret.Value.GetPlain()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if current == value {
			continue
		}
		secret.Value = entity.NewEncryptedField(value)
		if err = setting.SetData(secret); err != nil {
			return nil, hperrors.Wrap(err)
		}
		setting.UpdateVer++
		changed = append(changed, setting)
	}
	return changed, nil
}
