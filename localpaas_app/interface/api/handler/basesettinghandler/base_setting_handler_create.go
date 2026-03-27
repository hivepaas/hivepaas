package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudprovideruc/cloudproviderdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cronjobuc/cronjobdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc/healthcheckdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuilduc/imagebuilddto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc/imservicedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc/notificationdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc/oauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc/registryauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/repowebhookuc/repowebhookdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/ssluc/ssldto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

type CreateSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type CreateSettingOption func(*CreateSettingOptions)

func CreateSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) CreateSettingOption {
	return func(opts *CreateSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

//nolint:funlen,gocyclo
func (h *BaseSettingHandler) CreateSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.SettingScopeType,
	opts ...CreateSettingOption,
) {
	var auth *basedto.Auth
	var err error

	options := &CreateSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.SettingScope{}
	switch scopeType {
	case base.SettingScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "")
	case base.SettingScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "")
	case base.SettingScopeApp:
		auth, scope.ProjectID, scope.AppID, _, err = h.GetAuthAppSettings(ctx, base.ActionTypeWrite, "")
	case base.SettingScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "")
	default:
		err = apperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	var req any
	var ucFunc func() (any, error)
	reqCtx := h.RequestCtx(ctx)

	switch resType { //nolint:exhaustive
	case base.ResourceTypeBasicAuth:
		r := basicauthdto.NewCreateBasicAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.BasicAuthUC.CreateBasicAuth(reqCtx, auth, r) }

	case base.ResourceTypeGithubApp:
		r := githubappdto.NewCreateGithubAppReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.GithubAppUC.CreateGithubApp(reqCtx, auth, r) }

	case base.ResourceTypeAccessToken:
		r := accesstokendto.NewCreateAccessTokenReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AccessTokenUC.CreateAccessToken(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewCreateOAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.CreateOAuth(reqCtx, auth, r) }

	case base.ResourceTypeRegistryAuth:
		r := registryauthdto.NewCreateRegistryAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.RegistryAuthUC.CreateRegistryAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudProvider:
		r := cloudproviderdto.NewCreateCloudProviderReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CloudProviderUC.CreateCloudProvider(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewCreateCloudStorageReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.CreateCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeSSHKey:
		r := sshkeydto.NewCreateSSHKeyReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSHKeyUC.CreateSSHKey(reqCtx, auth, r) }

	case base.ResourceTypeSSL:
		r := ssldto.NewCreateSSLReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSLUC.CreateSSL(reqCtx, auth, r) }

	case base.ResourceTypeCronJob:
		r := cronjobdto.NewCreateCronJobReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CronJobUC.CreateCronJob(reqCtx, auth, r) }

	case base.ResourceTypeHealthcheck:
		r := healthcheckdto.NewCreateHealthcheckReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.HealthcheckUC.CreateHealthcheck(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewCreateSecretReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SecretUC.CreateSecret(reqCtx, auth, r) }

	case base.ResourceTypeAPIKey:
		r := apikeydto.NewCreateAPIKeyReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.APIKeyUC.CreateAPIKey(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewCreateIMServiceReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.CreateIMService(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewCreateEmailReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.EmailUC.CreateEmail(reqCtx, auth, r) }

	case base.ResourceTypeRepoWebhook:
		r := repowebhookdto.NewCreateRepoWebhookReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.RepoWebhookUC.CreateRepoWebhook(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewCreateNotificationReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.CreateNotification(reqCtx, auth, r) }

	case base.ResourceTypeImageBuild:
		r := imagebuilddto.NewCreateImageBuildReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ImageBuildUC.CreateImageBuild(reqCtx, auth, r) }

	case base.ResourceTypeFile:
		// NOTE: not implemented
		err = apperrors.NewNotImplementedNT()
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	if options.PreRequestHandler != nil {
		if err = options.PreRequestHandler(auth, req); err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	resp, err := ucFunc()
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}
