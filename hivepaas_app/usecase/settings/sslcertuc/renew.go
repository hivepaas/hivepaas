package sslcertuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

// RenewSSLCert obtains the certificate again and saves it as an update does:
// the apps that mount it follow it, and Traefik's files are written from what
// is stored.
func (uc *UC) RenewSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.RenewSSLCertReq,
) (*sslcertdto.RenewSSLCertResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID != setting.CurrentObjectID {
		return nil, hperrors.Wrap(hperrors.ErrInheritedSettingNonUpdatable)
	}

	_, err = uc.UpdateSetting(ctx, &settings.UpdateSettingReq{
		BaseSettingReq: req.BaseSettingReq,
		ID:             req.ID,
		Inheritable:    setting.Inheritable,
		Default:        setting.Default,
		UpdateVer:      setting.UpdateVer,
	}, &settings.UpdateSettingData{
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			_ *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			return uc.renewInto(ctx, db, req.Scope, pData.Setting)
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sslcertdto.RenewSSLCertResp{}, nil
}

// renewInto obtains the certificate anew into the setting being saved, and
// writes Traefik's files from it.
func (uc *UC) renewInto(
	ctx context.Context,
	db database.Tx,
	scope *entity.ObjectScope,
	setting *entity.Setting,
) error {
	// The setting is a copy of the one loaded, sharing its parsed data: the
	// certificate is obtained into data of its own, leaving the loaded one as
	// the record of what was there before.
	cert, err := setting.AsSSLCert()
	if err != nil {
		return hperrors.Wrap(err)
	}
	renewed := *cert
	if err = setting.SetData(&renewed); err != nil {
		return hperrors.Wrap(err)
	}

	refObjects := entity.NewRefObjects()
	if err := uc.SettingService.LoadRefObjects(ctx, db, &refObjects, scope, true, setting); err != nil {
		return hperrors.Wrap(err)
	}
	if _, err := uc.sslService.ObtainCert(ctx, setting, refObjects, false); err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(uc.sslService.WriteCertFiles(true, setting))
}
