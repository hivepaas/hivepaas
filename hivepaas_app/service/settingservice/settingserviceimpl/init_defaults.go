package settingserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func (s *service) InitDefaults(
	ctx context.Context,
	db database.IDB,
) (err error) {
	settings, _, err := s.settingRepo.List(ctx, db, base.NewObjectScopeGlobal(), nil,
		bunex.SelectColumns("id", "type", "status"),
	)
	if err != nil {
		return apperrors.New(err)
	}

	timeNow := timeutil.NowUTC()

	// Image build settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeImageBuildSettings
	}) {
		err = s.initDefaultImageBuildSettings(ctx, db, timeNow)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// Notification settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeNotification
	}) {
		err = s.initDefaultNotificationSettings(ctx, db, timeNow)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// Domain settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeDomainSettings
	}) {
		err = s.initDefaultDomainSettings(ctx, db, timeNow)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// Storage settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeStorageSettings
	}) {
		err = s.initDefaultStorageSettings(ctx, db, timeNow)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// System cleanup settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSystemCleanup
	}) {
		err = s.initDefaultSystemCleanup(ctx, db, timeNow)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// System backup settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSystemBackup
	}) {
		err = s.initDefaultSystemBackup(ctx, db, timeNow)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// SSL renewal settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSSLRenewal
	}) {
		err = s.initDefaultSSLRenewal(ctx, db, timeNow)
		if err != nil {
			return apperrors.New(err)
		}
	}

	// Default self-signed SSL cert
	err = s.initDefaultSSLSelfSigned(ctx, db, timeNow)
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
