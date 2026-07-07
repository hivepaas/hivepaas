package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc/emaildto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/healthcheckuc/healthcheckdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/imserviceuc/imservicedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/notificationuc/notificationdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/oauthuc/oauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/schedjobuc/schedjobdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sshkeyuc/sshkeydto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc/sslproviderdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/usersettings/apikeyuc/apikeydto"
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
func (h *Handler) CreateSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...CreateSettingOption,
) {
	var auth *basedto.Auth
	var err error

	options := &CreateSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "")
	case base.ObjectScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.AppID, _, err = h.GetAuthAppSettings(ctx, base.ActionTypeWrite, "")
	case base.ObjectScopeUser:
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
	case base.ResourceTypeAccessToken:
		r := accesstokendto.NewCreateAccessTokenReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AccessTokenUC.CreateAccessToken(reqCtx, auth, r) }

	case base.ResourceTypeAcmeDnsProvider:
		r := acmednsproviderdto.NewCreateAcmeDnsProviderReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AcmeDnsProviderUC.CreateAcmeDnsProvider(reqCtx, auth, r) }

	case base.ResourceTypeAPIKey:
		r := apikeydto.NewCreateAPIKeyReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.APIKeyUC.CreateAPIKey(reqCtx, auth, r) }

	case base.ResourceTypeBasicAuth:
		r := basicauthdto.NewCreateBasicAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.BasicAuthUC.CreateBasicAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewCreateCloudStorageReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.CreateCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeClusterNetwork:
		r := networkdto.NewCreateNetworkReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ClusterNetworkUC.CreateNetwork(reqCtx, auth, r) }

	case base.ResourceTypeClusterVolume:
		r := volumedto.NewCreateVolumeReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ClusterVolumeUC.CreateVolume(reqCtx, auth, r) }

	case base.ResourceTypeCommandTemplate:
		r := commandtemplatedto.NewCreateCommandTemplateReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CommandTemplateUC.CreateCommandTemplate(reqCtx, auth, r) }

	case base.ResourceTypeConfigFile:
		r := configfiledto.NewCreateConfigFileReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ConfigFileUC.CreateConfigFile(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewCreateEmailReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.EmailUC.CreateEmail(reqCtx, auth, r) }

	case base.ResourceTypeGithubApp:
		r := githubappdto.NewCreateGithubAppReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.GithubAppUC.CreateGithubApp(reqCtx, auth, r) }

	case base.ResourceTypeHealthcheck:
		r := healthcheckdto.NewCreateHealthcheckReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.HealthcheckUC.CreateHealthcheck(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewCreateIMServiceReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.CreateIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewCreateNotificationReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.CreateNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewCreateOAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.CreateOAuth(reqCtx, auth, r) }

	case base.ResourceTypeRegistryAuth:
		r := registryauthdto.NewCreateRegistryAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.RegistryAuthUC.CreateRegistryAuth(reqCtx, auth, r) }

	case base.ResourceTypeRepoWebhook:
		r := repowebhookdto.NewCreateRepoWebhookReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.RepoWebhookUC.CreateRepoWebhook(reqCtx, auth, r) }

	case base.ResourceTypeSchedJob:
		r := schedjobdto.NewCreateSchedJobReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SchedJobUC.CreateSchedJob(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewCreateSecretReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SecretUC.CreateSecret(reqCtx, auth, r) }

	case base.ResourceTypeSSHKey:
		r := sshkeydto.NewCreateSSHKeyReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSHKeyUC.CreateSSHKey(reqCtx, auth, r) }

	case base.ResourceTypeSSLCert:
		r := sslcertdto.NewCreateSSLCertReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSLCertUC.CreateSSLCert(reqCtx, auth, r) }

	case base.ResourceTypeSSLProvider:
		r := sslproviderdto.NewCreateSSLProviderReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSLProviderUC.CreateSSLProvider(reqCtx, auth, r) }
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
