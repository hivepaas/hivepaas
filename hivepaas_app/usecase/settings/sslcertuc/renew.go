package sslcertuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (uc *UC) RenewSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.RenewSSLCertReq,
) (*sslcertdto.RenewSSLCertResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	if setting.ObjectID != setting.CurrentObjectID {
		return nil, hperrors.Wrap(hperrors.ErrInheritedSettingNonUpdatable)
	}

	_, err = uc.sslService.ObtainCert(ctx, setting, resp.RefObjects, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &sslcertdto.RenewSSLCertResp{}, nil
}
