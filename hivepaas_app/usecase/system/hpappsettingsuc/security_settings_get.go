package hpappsettingsuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/hpappsettingsuc/hpappsettingsdto"
)

func (uc *UC) GetSecuritySettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *hpappsettingsdto.GetSecuritySettingsReq,
) (*hpappsettingsdto.GetSecuritySettingsResp, error) {
	input := &hpappsettingsdto.SecuritySettingsTransformInput{
		Config: config.Current(),
	}

	resp, err := hpappsettingsdto.TransformSecuritySettings(input)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &hpappsettingsdto.GetSecuritySettingsResp{
		Data: resp,
	}, nil
}
