package secretuc

import (
	"bytes"
	"context"
	"io"
	"net/url"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

func (uc *UC) DownloadSecret(
	ctx context.Context,
	auth *basedto.Auth,
	req *secretdto.DownloadSecretReq,
) (*secretdto.DownloadSecretResp, error) {
	tokenClaims, err := uc.FileService.ParseDownloadToken(req.Token)
	if err != nil {
		return nil, apperrors.Wrap(apperrors.ErrTokenInvalid).WithMsgLog("failed to parse download token")
	}

	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if resp.Data.ID != tokenClaims.FileID {
		return nil, apperrors.Wrap(apperrors.ErrTokenInvalid).
			WithMsgLog("setting ID mismatches the ID in the token")
	}

	secret, err := resp.Data.AsSecret()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	data, err := secret.ValueAsBytes()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	contentType := gofn.If(secret.Base64, "application/octet-stream", "text/plain")
	extraHeaders := map[string]string{
		"Content-Disposition": gofn.If(req.ViewInline, "inline; ", "attachment; ") +
			`filename*=UTF-8''` + url.QueryEscape(secret.Key),
	}

	return &secretdto.DownloadSecretResp{
		Data: &secretdto.DownloadSecretDataResp{
			BaseDownloadDataResp: &settings.BaseDownloadDataResp{
				ContentType:   contentType,
				ContentLength: int64(len(data)),
				ExtraHeaders:  extraHeaders,
				Content:       io.NopCloser(bytes.NewReader(data)),
			},
		},
	}, nil
}
