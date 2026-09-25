package settingmountserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) ClaimedPaths(
	ctx context.Context, db database.IDB, appID, exceptSettingID string,
) (map[string]string, error) {
	settings, err := s.loadClaimants(ctx, db, appID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	claimed := map[string]string{}
	for _, setting := range settings {
		if setting.ID == exceptSettingID || setting.Status != base.SettingStatusActive ||
			setting.Type != base.SettingTypeAppSettingMount {
			continue
		}
		mount, err := setting.AsAppSettingMount()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		for _, f := range mount.Files {
			if f != nil {
				claimed[f.Path] = "setting mount " + setting.Name
			}
		}
	}
	return claimed, nil
}

// loadClaimantsFromRepo is the app's own entries: the only settings that give
// it files.
func (s *service) loadClaimantsFromRepo(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.type = ?", base.SettingTypeAppSettingMount),
	)
	return settings, hperrors.Wrap(err)
}
