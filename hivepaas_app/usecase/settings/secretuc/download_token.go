package secretuc

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

const (
	defaultDownloadTokenExp    = 30 * time.Second
	defaultDownloadTokenExpDev = 60 * time.Second
)

func (uc *UC) GetDownloadToken(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.GetDownloadTokenReq,
) (*secretdto.GetDownloadTokenResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, uc.DB, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = uc.authorizeDownload(ctx, auth, req.Scope, resp.Data); err != nil {
		return nil, hperrors.Wrap(err)
	}

	expiration := req.Expiration.ToDuration()
	if expiration <= 0 {
		expiration = gofn.If(config.Current().IsDevEnv(), defaultDownloadTokenExpDev, defaultDownloadTokenExp)
	}
	token, err := uc.FileService.GenerateDownloadToken(auth.User.ID, resp.Data.ID, false, expiration)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &secretdto.GetDownloadTokenResp{
		Data: &secretdto.GetDownloadTokenDataResp{
			Token: token,
		},
	}, nil
}

// authorizeDownload lets a secret be downloaded only by whoever may reveal it: a
// download hands the value over in the clear, as a reveal does, so it takes the
// same gate - the operator's switch, the capability - and leaves the same
// record. An inherited secret belongs to the scope that made it, and is not
// revealed from below.
func (uc *UC) authorizeDownload(
	ctx context.Context,
	auth *basedto.Auth,
	scope *entity.ObjectScope,
	setting *entity.Setting,
) error {
	if setting.ObjectID != setting.CurrentObjectID {
		return hperrors.Wrap(hperrors.ErrUserNotHavePermissionOnRevealSecrets).
			WithMsgLog("an inherited secret is not downloaded from below")
	}
	return hperrors.Wrap(uc.AuthorizeReveal(ctx, uc.DB, auth, scope, setting))
}
