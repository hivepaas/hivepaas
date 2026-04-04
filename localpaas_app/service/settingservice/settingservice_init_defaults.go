package settingservice

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

func (s *settingService) InitDefaults(
	ctx context.Context,
	db database.IDB,
) (err error) {
	settings, _, err := s.settingRepo.List(ctx, db, base.NewSettingScopeGlobal(), nil,
		bunex.SelectWhereIn("setting.type IN (?)", base.SettingTypeImageBuild,
			base.SettingTypeSystemCleanup, base.SettingTypeSystemBackup, base.SettingTypeSSLRenewal),
		bunex.SelectExcludeColumns("data"),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()

	// Image build settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeImageBuild
	}) {
		err = s.initDefaultImageBuild(ctx, db, timeNow)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// System cleanup settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSystemCleanup
	}) {
		err = s.initDefaultSystemCleanup(ctx, db, timeNow)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// System backup settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSystemBackup
	}) {
		err = s.initDefaultSystemBackup(ctx, db, timeNow)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	// SSL renewal settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSSLRenewal
	}) {
		err = s.initDefaultSSLRenewal(ctx, db, timeNow)
		if err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}
