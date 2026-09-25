package settingmountserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
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
		if setting.ID == exceptSettingID || setting.Status != base.SettingStatusActive {
			continue
		}
		switch setting.Type { //nolint:exhaustive
		case base.SettingTypeSecret:
			secret, err := setting.AsSecret()
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			if target := settingmountservice.SecretFileTarget(secret); target != "" {
				claimed[target] = "secret " + setting.Name
			}
		case base.SettingTypeConfigFile:
			configFile, err := setting.AsConfigFile()
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			if target := settingmountservice.ConfigFileTarget(configFile); target != "" {
				claimed[target] = "config file " + setting.Name
			}
		case base.SettingTypeAppSettingMount:
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
	}
	return claimed, nil
}

// loadClaimantsFromRepo is the app's own settings that can give it files.
func (s *service) loadClaimantsFromRepo(ctx context.Context, db database.IDB, appID string) ([]*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.object_id = ?", appID),
		bunex.SelectWhere("setting.type IN (?)", bunex.List([]base.SettingType{
			base.SettingTypeSecret, base.SettingTypeConfigFile, base.SettingTypeAppSettingMount})),
	)
	return settings, hperrors.Wrap(err)
}
