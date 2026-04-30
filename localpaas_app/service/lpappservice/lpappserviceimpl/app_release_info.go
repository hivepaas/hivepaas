package lpappserviceimpl

import (
	"context"
	"encoding/json"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/httputil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/version"
	"github.com/localpaas/localpaas/localpaas_app/service/lpappservice"
)

const (
	urlAppReleaseInfo = "https://raw.githubusercontent.com/localpaas/localpaas/main/release.json"
)

func (s *service) GetAppReleaseInfo(ctx context.Context) (*lpappservice.AppReleaseInfo, error) {
	data, err := httputil.HTTPGet(ctx, urlAppReleaseInfo)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	info := &lpappservice.AppReleaseInfo{}
	err = json.Unmarshal(data, info)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if info.Stable != nil && info.Stable.AppVersion != "" {
		cmp, err := version.CmpStr(info.Stable.AppVersion, base.StableVersion.AppVersion)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		info.Stable.CanUpdate = cmp > 0
	}

	if info.Beta != nil && info.Beta.AppVersion != "" {
		cmp, err := version.CmpStr(info.Beta.AppVersion, base.BetaVersion.AppVersion)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		info.Beta.CanUpdate = cmp > 0
	}

	return info, nil
}
