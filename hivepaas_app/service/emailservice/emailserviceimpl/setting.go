package emailserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
)

func (s *service) GetDefaultSystemEmail(
	ctx context.Context,
	db database.IDB,
) (*entity.Setting, error) {
	settings, _, err := s.settingRepo.List(ctx, db, base.NewObjectScopeGlobal(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeEmail),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectOrder("setting.is_default DESC"),
		bunex.SelectLimit(2), //nolint
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if (len(settings) == 1 || (len(settings) > 1 && settings[0].Default)) && settings[0].IsActive() {
		return settings[0], nil
	}
	return nil, apperrors.NewNotFound("Email setting").
		WithMsgLog("default system email setting not found")
}
