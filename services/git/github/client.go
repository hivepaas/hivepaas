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
	accessToken    string

	appsTransport    *ghinstallation.AppsTransport
	installTransport *ghinstallation.Transport

	client    *gogithub.Client
	appClient *gogithub.Client
}

func (c *Client) IsAppClient() bool {
	return c.appID > 0
}

func (c *Client) IsTokenClient() bool {
	return c.accessToken != ""
}

func (c *Client) CreateAppToken(ctx context.Context) (string, error) {
	if !c.IsAppClient() {
		return "", apperrors.Wrap(ErrGithubAppClientRequired)
	}
	token, err := c.installTransport.Token(ctx)
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	return token, nil
}

func NewFromApp(appID, installationID int64, privateKey []byte) (*Client, error) {
	appTr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, appID, privateKey)
	if err != nil {
		return nil, apperrors.Wrap(err)
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

func NewFromToken(accessToken string) (*Client, error) {
	client := &Client{
		accessToken: accessToken,
		client: gogithub.NewClient(&http.Client{
			Transport: NewPatTransport(http.DefaultTransport, accessToken),
		}),
	}
	return client, nil
}

func NewFromSetting(setting *entity.Setting) (*Client, error) {
	switch setting.Type { //nolint:exhaustive
	case base.SettingTypeGithubApp:
		githubApp, err := setting.AsGithubApp()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		privateKey, err := githubApp.PrivateKey.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return NewFromApp(githubApp.AppID, githubApp.InstallationID, reflectutil.UnsafeStrToBytes(privateKey))

	case base.SettingTypeAccessToken:
		if base.AccessTokenKind(setting.Kind) != base.AccessTokenKindGithub {
			return nil, apperrors.Wrap(ErrAccessProviderInvalid).
				WithMsgLog("token kind '%s' is unsupported", setting.Kind)
		}
		gitToken, err := setting.AsAccessToken()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		token, err := gitToken.Token.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		return NewFromToken(token)

	default:
		return nil, apperrors.Wrap(ErrAccessProviderInvalid).
			WithMsgLog("setting type '%s' is invalid", setting.Type)
	}
}
