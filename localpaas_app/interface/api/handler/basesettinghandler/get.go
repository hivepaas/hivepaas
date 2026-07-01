package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/configfileuc/configfiledto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/healthcheckuc/healthcheckdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc/imservicedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/notificationuc/notificationdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc/oauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/registryauthuc/registryauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/repowebhookuc/repowebhookdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/schedjobuc/schedjobdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc/sslcertdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslprovideruc/sslproviderdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

type GetSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type GetSettingOption func(*GetSettingOptions)

func GetSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) GetSettingOption {
	return func(opts *GetSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

//nolint:funlen,gocyclo
func (h *Handler) GetSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...GetSettingOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &GetSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeRead, "itemID")
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx, base.ActionTypeRead, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeRead, "itemID")
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
		r := accesstokendto.NewGetAccessTokenReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.AccessTokenUC.GetAccessToken(reqCtx, auth, r) }

	case base.ResourceTypeAcmeDnsProvider:
		r := acmednsproviderdto.NewGetAcmeDnsProviderReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.AcmeDnsProviderUC.GetAcmeDnsProvider(reqCtx, auth, r) }

	case base.ResourceTypeAPIKey:
		r := apikeydto.NewGetAPIKeyReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.APIKeyUC.GetAPIKey(reqCtx, auth, r) }

	case base.ResourceTypeBasicAuth:
		r := basicauthdto.NewGetBasicAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.BasicAuthUC.GetBasicAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewGetCloudStorageReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.GetCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeClusterVolume:
		r := volumedto.NewGetVolumeReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ClusterVolumeUC.GetVolume(reqCtx, auth, r) }

	case base.ResourceTypeConfigFile:
		r := configfiledto.NewGetConfigFileReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ConfigFileUC.GetConfigFile(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewGetEmailReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.EmailUC.GetEmail(reqCtx, auth, r) }

	case base.ResourceTypeGithubApp:
		r := githubappdto.NewGetGithubAppReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.GithubAppUC.GetGithubApp(reqCtx, auth, r) }

	case base.ResourceTypeHealthcheck:
		r := healthcheckdto.NewGetHealthcheckReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.HealthcheckUC.GetHealthcheck(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewGetIMServiceReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.GetIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewGetNotificationReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.GetNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewGetOAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.GetOAuth(reqCtx, auth, r) }

	case base.ResourceTypeRegistryAuth:
		r := registryauthdto.NewGetRegistryAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RegistryAuthUC.GetRegistryAuth(reqCtx, auth, r) }

	case base.ResourceTypeRepoWebhook:
		r := repowebhookdto.NewGetRepoWebhookReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RepoWebhookUC.GetRepoWebhook(reqCtx, auth, r) }

	case base.ResourceTypeSchedJob:
		r := schedjobdto.NewGetSchedJobReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SchedJobUC.GetSchedJob(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewGetSecretReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SecretUC.GetSecret(reqCtx, auth, r) }

	case base.ResourceTypeSSHKey:
		r := sshkeydto.NewGetSSHKeyReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSHKeyUC.GetSSHKey(reqCtx, auth, r) }

	case base.ResourceTypeSSLCert:
		r := sslcertdto.NewGetSSLCertReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSLCertUC.GetSSLCert(reqCtx, auth, r) }

	case base.ResourceTypeSSLProvider:
		r := sslproviderdto.NewGetSSLProviderReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSLProviderUC.GetSSLProvider(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = apperrors.NewNotImplementedNT()
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
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

	ctx.JSON(http.StatusOK, resp)
}
