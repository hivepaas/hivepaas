package sessionuc

import (
	"context"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc/sessiondto"
)

var (
	mapProviderIcon = map[base.OAuthKind]string{
		base.OAuthKindGithubApp: "github", // icon mapping for special oauth types only
	}
)

func (uc *UC) GetLoginOptions(
	ctx context.Context,
	req *sessiondto.GetLoginOptionsReq,
) (*sessiondto.GetLoginOptionsResp, error) {
	settings, _, err := uc.settingRepo.List(ctx, uc.db, base.NewSettingScopeGlobal(), nil,
		bunex.SelectWhere("setting.type = ? OR setting.type = ?",
			base.SettingTypeOAuth, base.SettingTypeGithubApp),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectOrder("kind"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	var resp []*sessiondto.LoginOptionResp
	for _, setting := range settings {
		if setting.Type == base.SettingTypeGithubApp {
			app, err := setting.AsGithubApp()
			if err != nil || !app.SSOEnabled {
				continue
			}
		}

		oauthType := base.OAuthKind(setting.Kind)
		resp = append(resp, &sessiondto.LoginOptionResp{
			Type:    oauthType,
			Name:    setting.Name,
			Icon:    gofn.Coalesce(mapProviderIcon[oauthType], string(oauthType)),
			AuthURL: "/_/auth/sso/" + setting.ID,
		})
	}

	return &sessiondto.GetLoginOptionsResp{
		Data: resp,
	}, nil
}
