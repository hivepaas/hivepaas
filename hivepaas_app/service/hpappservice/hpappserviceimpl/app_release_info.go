package hpappserviceimpl

import (
	"context"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/httputil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/version"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	urlAppReleaseInfo = "https://raw.githubusercontent.com/hivepaas/hivepaas/main/release.json"
)

func (s *service) GetAppReleaseInfo(ctx context.Context) (*hpappservice.AppReleaseInfo, error) {
	data, err := httputil.HTTPGet(ctx, urlAppReleaseInfo)
	if err != nil {
		return nil, apperrors.New(err)
	}

	info := &hpappservice.AppReleaseInfo{}
	err = json.Unmarshal(data, info)
	if err != nil {
		return nil, apperrors.New(err)
	}

	if info.Stable != nil && info.Stable.AppVersion != "" {
		cmp, err := version.CmpStr(info.Stable.AppVersion, base.StableVersion.AppVersion)
		if err != nil {
			return nil, apperrors.New(err)
		}
		info.Stable.CanUpdate = cmp > 0
	}

	if info.Beta != nil && info.Beta.AppVersion != "" {
		cmp, err := version.CmpStr(info.Beta.AppVersion, base.BetaVersion.AppVersion)
		if err != nil {
			return nil, apperrors.New(err)
		}
		info.Beta.CanUpdate = cmp > 0
	}

	return info, nil
}
