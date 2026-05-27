package basesettinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc/accesstokendto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/configfileuc/configfiledto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/fileuc/filedto"
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
	"github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
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

//nolint:funlen
func (h *Handler) DeleteSetting(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.SettingScopeType,
	opts ...DeleteSettingOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &DeleteSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.SettingScope{}
	switch scopeType {
	case base.SettingScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeDelete, "itemID")
	case base.SettingScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.SettingScopeApp:
		auth, scope.ProjectID, scope.AppID, itemID, err = h.GetAuthAppSettings(ctx, base.ActionTypeWrite, "itemID")
	case base.SettingScopeUser:
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
	case base.ResourceTypeBasicAuth:
		r := basicauthdto.NewDeleteBasicAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.BasicAuthUC.DeleteBasicAuth(reqCtx, auth, r) }

	case base.ResourceTypeGithubApp:
		r := githubappdto.NewDeleteGithubAppReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.GithubAppUC.DeleteGithubApp(reqCtx, auth, r) }

	case base.ResourceTypeAccessToken:
		r := accesstokendto.NewDeleteAccessTokenReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.AccessTokenUC.DeleteAccessToken(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewDeleteOAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.DeleteOAuth(reqCtx, auth, r) }

	case base.ResourceTypeRegistryAuth:
		r := registryauthdto.NewDeleteRegistryAuthReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RegistryAuthUC.DeleteRegistryAuth(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewDeleteCloudStorageReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.DeleteCloudStorage(reqCtx, auth, r) }

	case base.ResourceTypeSSHKey:
		r := sshkeydto.NewDeleteSSHKeyReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSHKeyUC.DeleteSSHKey(reqCtx, auth, r) }

	case base.ResourceTypeSSLCert:
		r := sslcertdto.NewDeleteSSLCertReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSLCertUC.DeleteSSLCert(reqCtx, auth, r) }

	case base.ResourceTypeSchedJob:
		r := schedjobdto.NewDeleteSchedJobReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SchedJobUC.DeleteSchedJob(reqCtx, auth, r) }

	case base.ResourceTypeHealthcheck:
		r := healthcheckdto.NewDeleteHealthcheckReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.HealthcheckUC.DeleteHealthcheck(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewDeleteSecretReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SecretUC.DeleteSecret(reqCtx, auth, r) }

	case base.ResourceTypeConfigFile:
		r := configfiledto.NewDeleteConfigFileReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ConfigFileUC.DeleteConfigFile(reqCtx, auth, r) }

	case base.ResourceTypeAPIKey:
		r := apikeydto.NewDeleteAPIKeyReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.APIKeyUC.DeleteAPIKey(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewDeleteIMServiceReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.DeleteIMService(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewDeleteEmailReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.EmailUC.DeleteEmail(reqCtx, auth, r) }

	case base.ResourceTypeRepoWebhook:
		r := repowebhookdto.NewDeleteRepoWebhookReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RepoWebhookUC.DeleteRepoWebhook(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewDeleteNotificationReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.DeleteNotification(reqCtx, auth, r) }

	case base.ResourceTypeFile:
		r := filedto.NewDeleteFileReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.FileUC.DeleteFile(reqCtx, auth, r) }

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
