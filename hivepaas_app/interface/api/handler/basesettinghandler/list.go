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

type ListSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type ListSettingOption func(*ListSettingOptions)

func ListSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) ListSettingOption {
	return func(opts *ListSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

//nolint:funlen,gocyclo
func (h *Handler) ListSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...ListSettingOption,
) {
	var auth *basedto.Auth
	var err error

	options := &ListSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeRead, "")
	case base.ObjectScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectSettings(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.AppID, _, err = h.GetAuthAppSettings(ctx, base.ActionTypeRead, "")
	case base.ObjectScopeUser:
		auth, scope.UserID, _, err = h.GetAuthUserSettings(ctx, base.ActionTypeRead, "")
	default:
		err = apperrors.NewUnsupported("Setting scope 'none'")
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	var req any
	var paging *basedto.Paging
	var ucFunc func() (any, error)
	reqCtx := h.RequestCtx(ctx)

	switch resType { //nolint:exhaustive
	case base.ResourceTypeAccessToken:
		r := accesstokendto.NewListAccessTokenReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AccessTokenUC.ListAccessToken(reqCtx, auth, r) }

	case base.ResourceTypeAcmeDnsProvider:
		r := acmednsproviderdto.NewListAcmeDnsProviderReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.AcmeDnsProviderUC.ListAcmeDnsProvider(reqCtx, auth, r) }

	case base.ResourceTypeAPIKey:
		r := apikeydto.NewListAPIKeyReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.APIKeyUC.ListAPIKey(reqCtx, auth, r) }

	case base.ResourceTypeBasicAuth:
		r := basicauthdto.NewListBasicAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.BasicAuthUC.ListBasicAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewListCloudStorageReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.ListCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeClusterNetwork:
		r := networkdto.NewListNetworkReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ClusterNetworkUC.ListNetwork(reqCtx, auth, r) }

	case base.ResourceTypeClusterVolume:
		r := volumedto.NewListVolumeReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ClusterVolumeUC.ListVolume(reqCtx, auth, r) }

	case base.ResourceTypeConfigFile:
		r := configfiledto.NewListConfigFileReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.ConfigFileUC.ListConfigFile(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewListEmailReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.EmailUC.ListEmail(reqCtx, auth, r) }

	case base.ResourceTypeGithubApp:
		r := githubappdto.NewListGithubAppReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.GithubAppUC.ListGithubApp(reqCtx, auth, r) }

	case base.ResourceTypeHealthcheck:
		r := healthcheckdto.NewListHealthcheckReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.HealthcheckUC.ListHealthcheck(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewListIMServiceReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.ListIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewListNotificationReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.ListNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewListOAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.ListOAuth(reqCtx, auth, r) }

	case base.ResourceTypeRegistryAuth:
		r := registryauthdto.NewListRegistryAuthReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.RegistryAuthUC.ListRegistryAuth(reqCtx, auth, r) }

	case base.ResourceTypeRepoWebhook:
		r := repowebhookdto.NewListRepoWebhookReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.RepoWebhookUC.ListRepoWebhook(reqCtx, auth, r) }

	case base.ResourceTypeSchedJob:
		r := schedjobdto.NewListSchedJobReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SchedJobUC.ListSchedJob(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewListSecretReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SecretUC.ListSecret(reqCtx, auth, r) }

	case base.ResourceTypeSSHKey:
		r := sshkeydto.NewListSSHKeyReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSHKeyUC.ListSSHKey(reqCtx, auth, r) }

	case base.ResourceTypeSSLCert:
		r := sslcertdto.NewListSSLCertReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSLCertUC.ListSSLCert(reqCtx, auth, r) }

	case base.ResourceTypeSSLProvider:
		r := sslproviderdto.NewListSSLProviderReq()
		r.Scope = scope
		req, ucFunc = r, func() (any, error) { return h.SSLProviderUC.ListSSLProvider(reqCtx, auth, r) }

	default:
		// NOTE: not implemented
		err = apperrors.NewNotImplementedNT()
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateRequest(ctx, req, paging); err != nil {
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
