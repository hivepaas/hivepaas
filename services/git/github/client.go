package github

import (
	"context"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	gogithub "github.com/google/go-github/v85/github"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/reflectutil"
)

type Client struct {
	appID          int64
	installationID int64
	personalToken  string

	appsTransport    *ghinstallation.AppsTransport
	installTransport *ghinstallation.Transport

	client    *gogithub.Client
	appClient *gogithub.Client
}

func (c *Client) IsAppClient() bool {
	return c.appID > 0
}

func (c *Client) IsTokenClient() bool {
	return c.personalToken != ""
}

func (c *Client) CreateAppToken(ctx context.Context) (string, error) {
	if !c.IsAppClient() {
		return "", apperrors.New(ErrGithubAppClientRequired)
	}
	token, err := c.installTransport.Token(ctx)
	if err != nil {
		return "", apperrors.New(err)
	}
	return token, nil
}

func NewFromApp(appID, installationID int64, privateKey []byte) (*Client, error) {
	appTr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, privateKey)
	if err != nil {
		return nil, apperrors.New(err)
	}

	appClient := gogithub.NewClient(&http.Client{Transport: appTr})

	client := &Client{
		appID:          appID,
		installationID: installationID,
		appsTransport:  appTr,
		appClient:      appClient,
	}
	if installationID != 0 {
		client.installTransport = ghinstallation.NewFromAppsTransport(appTr, installationID)
		client.client = gogithub.NewClient(&http.Client{Transport: client.installTransport})
	} else {
		client.client = appClient
	}

	return client, nil
}

func NewFromPersonalToken(personalToken string) (*Client, error) {
	client := &Client{
		personalToken: personalToken,
		client: gogithub.NewClient(&http.Client{
			Transport: NewPatTransport(http.DefaultTransport, personalToken),
		}),
	}
	return client, nil
}

func NewFromSetting(setting *entity.Setting) (*Client, error) {
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		githubApp, err := setting.AsGithubApp()
		if err != nil {
			return nil, apperrors.New(err)
		}
		privateKey, err := githubApp.PrivateKey.GetPlain()
		if err != nil {
			return nil, apperrors.New(err)
		}
		return NewFromApp(githubApp.AppID, githubApp.InstallationID, reflectutil.UnsafeStrToBytes(privateKey))

	case base.SettingTypeAccessToken:
		if base.AccessTokenKind(setting.Kind) != base.AccessTokenKindGithub {
			return nil, apperrors.New(ErrAccessProviderInvalid).
				WithMsgLog("token kind '%s' is unsupported", setting.Kind)
		}
		gitToken, err := setting.AsAccessToken()
		if err != nil {
			return nil, apperrors.New(err)
		}
		token, err := gitToken.Token.GetPlain()
		if err != nil {
			return nil, apperrors.New(err)
		}
		return NewFromPersonalToken(token)

	default:
		return nil, apperrors.New(ErrAccessProviderInvalid).
			WithMsgLog("setting type '%s' is invalid", setting.Type)
	}
}
