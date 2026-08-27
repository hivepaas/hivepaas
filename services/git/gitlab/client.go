package gitlab

import (
	gogitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/httpclient"
)

type Client struct {
	token   string
	baseURL string

	client      *gogitlab.Client
	currentUser *gogitlab.User
}

func NewFromToken(token string, baseURL string) (*Client, error) {
	options := []gogitlab.ClientOptionFunc{
		gogitlab.WithHTTPClient(httpclient.DefaultClient),
	}
	if baseURL != "" {
		options = append(options, gogitlab.WithBaseURL(baseURL))
	}
	client, err := gogitlab.NewClient(token, options...)
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
		if tokenKind != base.AccessTokenKindGitlab {
			return nil, hperrors.Wrap(ErrAccessProviderInvalid).
				WithMsgLog("git source '%s' is invalid", setting.Kind)
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
