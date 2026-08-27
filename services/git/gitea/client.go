package gitea

import (
	gogitea "code.gitea.io/sdk/gitea"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type Client struct {
	token   string
	baseURL string

	client *gogitea.Client
}

func NewFromToken(token string, baseURL string) (*Client, error) {
	client, err := gogitea.NewClient(baseURL, gogitea.SetToken(token))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &Client{
		token:   token,
		baseURL: baseURL,
		client:  client,
	}, nil
}

func NewFromSetting(setting *entity.Setting) (*Client, error) {
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeAccessToken:
		gitToken, err := setting.AsAccessToken()
		tokenKind := base.AccessTokenKind(setting.Kind)
		if tokenKind != base.AccessTokenKindGitea {
			return nil, hperrors.Wrap(ErrAccessProviderInvalid).
				WithMsgLog("token kind '%s' is unsupported", tokenKind)
		}
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		token, err := gitToken.Token.GetPlain()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		return NewFromToken(token, gitToken.BaseURL)

	default:
		return nil, hperrors.Wrap(ErrAccessProviderInvalid)
	}
}
