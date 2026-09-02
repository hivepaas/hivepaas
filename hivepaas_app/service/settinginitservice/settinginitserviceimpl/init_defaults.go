package settinginitserviceimpl

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
)

const (
	lockIDSettingUpdate = "lock:sys:setting-update"
)

func (s *service) InitDefaults(
	ctx context.Context,
	db database.IDB,
) (err error) {
	err = transaction.Execute(ctx, db, func(db database.Tx) error {
		err = s.InitDefaultsWithTx(ctx, db)
		if err != nil {
			return hperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (s *service) InitDefaultsWithTx(
	ctx context.Context,
	db database.Tx,
) (err error) {
	// Acquire a lock for this process
	_, err = s.lockRepo.GetByID(ctx, db, lockIDSettingUpdate, bunex.SelectFor("UPDATE"))
	if err != nil {
		return hperrors.Wrap(err)
	}

	settings, _, err := s.settingRepo.List(ctx, db, entity.NewObjectScopeGlobal(), nil,
		bunex.SelectColumns("id", "type", "status"),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	timeNow := timeutil.NowUTC()

	// App placement settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeAppPlacement
	}) {
		err = s.initDefaultAppPlacementSettings(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Image build settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeImageBuild
	}) {
		err = s.initDefaultImageBuildSettings(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Notification settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeNotification
	}) {
		err = s.initDefaultNotificationSettings(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Domain settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeDomainSettings
	}) {
		err = s.initDefaultDomainSettings(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Storage settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeStorageSettings
	}) {
		err = s.initDefaultStorageSettings(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// System cleanup settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSystemCleanup
	}) {
		err = s.initDefaultSystemCleanup(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// System backup settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSystemBackup
	}) {
		err = s.initDefaultSystemBackup(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// SSL renewal settings
	if !gofn.ContainBy(settings, func(item *entity.Setting) bool {
		return item.Type == base.SettingTypeSSLRenewal
	}) {
		err = s.initDefaultSSLRenewal(ctx, db, timeNow)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	// Default self-signed SSL cert
	err = s.initDefaultSSLSelfSigned(ctx, db, timeNow)
	if err != nil {
		return hperrors.Wrap(err)
	}

	return nil
}
