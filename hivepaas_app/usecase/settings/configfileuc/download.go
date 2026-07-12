package configfileuc

import (
	"bytes"
	"context"
	"io"
	"net/url"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

func (uc *UC) DownloadConfigFile(
	ctx context.Context,
	auth *basedto.Auth,
	req *configfiledto.DownloadConfigFileReq,
) (*configfiledto.DownloadConfigFileResp, error) {
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

	configFile, err := resp.Data.AsConfigFile()
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	data := configFile.ContentAsBytes()
	contentType := gofn.If(configFile.Base64, "application/octet-stream", "text/plain")
	extraHeaders := map[string]string{
		"Content-Disposition": gofn.If(req.ViewInline, "inline; ", "attachment; ") +
			`filename*=UTF-8''` + url.QueryEscape(configFile.Name),
	}

	return &configfiledto.DownloadConfigFileResp{
		Data: &configfiledto.DownloadConfigFileDataResp{
			BaseDownloadDataResp: &settings.BaseDownloadDataResp{
				ContentType:   contentType,
				ContentLength: int64(len(data)),
				ExtraHeaders:  extraHeaders,
				Content:       io.NopCloser(bytes.NewReader(data)),
			},
		},
	}, nil
}
