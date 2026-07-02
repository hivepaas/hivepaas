package settingserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	domainSettingName            = "Domain settings"
	domainCertTypeDefault        = base.SSLCertTypeLetsEncrypt
	domainCertKeyTypeDefault     = base.SSLKeyTypeECP256
	domainCertValidPeriodDefault = timeutil.Day * 365 // For self-signed certs only
	domainCertAutoRenewDefault   = true
)

func (s *service) initDefaultDomainSettings(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	domainSetting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.ObjectScopeGlobal,
		Type:            base.SettingTypeDomainSettings,
		Status:          base.SettingStatusActive,
		Name:            domainSettingName,
		AvailInProjects: true,
		Default:         true,
		Version:         entity.CurrentDomainSettingsVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	domain := &entity.DomainSettings{
		RootDomain: config.Current.RootDomain,
		CertSettings: &entity.DomainCertSettings{
			CertType:    domainCertTypeDefault,
			KeyType:     domainCertKeyTypeDefault,
			ValidPeriod: timeutil.Duration(domainCertValidPeriodDefault),
			Email:       config.Current.Users.Admin.Email,
			AutoRenew:   domainCertAutoRenewDefault,
		},
	}
	domainSetting.MustSetData(domain)

	err = s.settingRepo.Insert(ctx, db, domainSetting)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
