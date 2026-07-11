package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
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

type DeleteSettingOptions struct {
	PreRequestHandler func(auth *basedto.Auth, req any) error
}

type DeleteSettingOption func(*DeleteSettingOptions)

func DeleteSettingPreRequestHandler(fn func(auth *basedto.Auth, req any) error) DeleteSettingOption {
	return func(opts *DeleteSettingOptions) {
		opts.PreRequestHandler = fn
	}
}

//nolint:funlen,gocyclo
func (h *Handler) DeleteSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.ObjectScopeType,
	opts ...DeleteSettingOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &DeleteSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.ObjectScope{}
	switch scopeType {
	case base.ObjectScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeDelete, "itemID")
	case base.ObjectScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeApp:
		auth, scope.ProjectID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.ObjectScopeUser:
		auth, scope.UserID, itemID, err = h.GetAuthUserSettings(ctx, base.ActionTypeWrite, "itemID")
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
		r := accesstokendto.NewDeleteAccessTokenReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.AccessTokenUC.DeleteAccessToken(reqCtx, auth, r) }

	case base.ResourceTypeAcmeDnsProvider:
		r := acmednsproviderdto.NewDeleteAcmeDnsProviderReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.AcmeDnsProviderUC.DeleteAcmeDnsProvider(reqCtx, auth, r) }

	case base.ResourceTypeAPIKey:
		r := apikeydto.NewDeleteAPIKeyReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.APIKeyUC.DeleteAPIKey(reqCtx, auth, r) }

	case base.ResourceTypeBasicAuth:
		r := basicauthdto.NewDeleteBasicAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.BasicAuthUC.DeleteBasicAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewDeleteCloudStorageReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.DeleteCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeClusterNetwork:
		r := networkdto.NewDeleteNetworkReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ClusterNetworkUC.DeleteNetwork(reqCtx, auth, r) }

	case base.ResourceTypeClusterNode:
		r := nodedto.NewDeleteNodeReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ClusterNodeUC.DeleteNode(reqCtx, auth, r) }

	case base.ResourceTypeClusterVolume:
		r := volumedto.NewDeleteVolumeReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ClusterVolumeUC.DeleteVolume(reqCtx, auth, r) }

	case base.ResourceTypeCommandTemplate:
		r := commandtemplatedto.NewDeleteCommandTemplateReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.CommandTemplateUC.DeleteCommandTemplate(reqCtx, auth, r) }

	case base.ResourceTypeConfigFile:
		r := configfiledto.NewDeleteConfigFileReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ConfigFileUC.DeleteConfigFile(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewDeleteEmailReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.EmailUC.DeleteEmail(reqCtx, auth, r) }

	case base.ResourceTypeGithubApp:
		r := githubappdto.NewDeleteGithubAppReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.GithubAppUC.DeleteGithubApp(reqCtx, auth, r) }

	case base.ResourceTypeHealthcheck:
		r := healthcheckdto.NewDeleteHealthcheckReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.HealthcheckUC.DeleteHealthcheck(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewDeleteIMServiceReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.DeleteIMService(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewDeleteNotificationReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.DeleteNotification(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewDeleteOAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.DeleteOAuth(reqCtx, auth, r) }

	case base.ResourceTypeRegistryAuth:
		r := registryauthdto.NewDeleteRegistryAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RegistryAuthUC.DeleteRegistryAuth(reqCtx, auth, r) }

	case base.ResourceTypeRepoWebhook:
		r := repowebhookdto.NewDeleteRepoWebhookReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RepoWebhookUC.DeleteRepoWebhook(reqCtx, auth, r) }

	case base.ResourceTypeSchedJob:
		r := schedjobdto.NewDeleteSchedJobReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SchedJobUC.DeleteSchedJob(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewDeleteSecretReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SecretUC.DeleteSecret(reqCtx, auth, r) }

	case base.ResourceTypeSSHKey:
		r := sshkeydto.NewDeleteSSHKeyReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSHKeyUC.DeleteSSHKey(reqCtx, auth, r) }

	case base.ResourceTypeSSLCert:
		r := sslcertdto.NewDeleteSSLCertReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSLCertUC.DeleteSSLCert(reqCtx, auth, r) }

	case base.ResourceTypeSSLProvider:
		r := sslproviderdto.NewDeleteSSLProviderReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSLProviderUC.DeleteSSLProvider(reqCtx, auth, r) }

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
