package registryserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

func (s *service) Sync(ctx context.Context, db database.IDB) (*registryservice.SyncResp, error) {
	resp := &registryservice.SyncResp{}
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeRegistry, false,
		bunex.SelectFor("UPDATE OF setting"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	wanted := false
	if setting != nil {
		cfg, err := setting.AsRegistrySettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		wanted = cfg.Enabled
	}

	app := &systemappservice.SyncedApp{Key: base.HivepaasRegistryKey, Name: registryAppName, Wanted: wanted}
	if err = systemappservice.RemoveIfServiceGone(ctx, db, s.systemAppService, app); err != nil {
		return nil, hperrors.Wrap(err)
	}

	switch {
	case setting == nil && app.Before != nil:
		// Never configured, or its row is gone: the app is not wanted, and
		// Apply would return before looking.
		if err = s.systemAppService.Remove(ctx, db, app.Before, false); err != nil {
			return nil, hperrors.Wrap(err)
		}
	case setting != nil:
		// The credential a removal names is left: it is deleted only from the
		// settings screen, which checks that no app still names it.
		applied, applyErr := s.Apply(ctx, db, &registryservice.SettingApplyReq{Setting: setting, RemoveApp: true})
		if applied != nil {
			resp.Cleanup = applied.Cleanup
			if applied.DeploymentTask != nil {
				resp.Tasks = append(resp.Tasks, applied.DeploymentTask)
			}
			resp.Tasks = append(resp.Tasks, applied.CertTasks...)
		}
		if applyErr != nil {
			return resp, hperrors.Wrap(applyErr)
		}
	}

	resp.App, _, err = systemappservice.SyncOutcome(ctx, db, s.systemAppService, app, resp.Tasks)
	return resp, hperrors.Wrap(err)
}
