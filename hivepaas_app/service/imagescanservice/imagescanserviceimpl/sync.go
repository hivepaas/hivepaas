package imagescanserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

func (s *service) Sync(ctx context.Context, db database.IDB) (*imagescanservice.SyncResp, error) {
	resp := &imagescanservice.SyncResp{}
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeImageScan, false,
		bunex.SelectFor("UPDATE OF setting"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	wanted := false
	if setting != nil {
		cfg, err := setting.AsImageScanSettings()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		wanted = cfg.Enabled
	}

	apps := []*systemappservice.SyncedApp{
		{Key: base.HivepaasTrivyKey, Name: serverAppName, Wanted: wanted},
		{Key: base.HivepaasTrivyRunnerKey, Name: runnerAppName, Wanted: wanted},
	}
	for _, app := range apps {
		if err = systemappservice.RemoveIfServiceGone(ctx, db, s.systemAppService, app); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	if setting == nil {
		// Never configured, or its row is gone: no app is wanted, and Apply
		// would return before looking. The runner goes first, as Apply has it.
		for i := len(apps) - 1; i >= 0; i-- {
			if app := apps[i]; app.Before != nil {
				if err = s.systemAppService.Remove(ctx, db, app.Before, false); err != nil {
					return nil, hperrors.Wrap(err)
				}
			}
		}
	} else {
		applied, applyErr := s.Apply(ctx, db, &imagescanservice.SettingApplyReq{Setting: setting})
		if applied != nil {
			resp.Tasks, resp.Cleanup = applied.Tasks, applied.Cleanup
		}
		if applyErr != nil {
			return resp, hperrors.Wrap(applyErr)
		}
	}

	for _, app := range apps {
		out, _, err := systemappservice.SyncOutcome(ctx, db, s.systemAppService, app, resp.Tasks)
		if err != nil {
			return resp, hperrors.Wrap(err)
		}
		if out != nil {
			resp.Apps = append(resp.Apps, out)
		}
	}
	return resp, nil
}
