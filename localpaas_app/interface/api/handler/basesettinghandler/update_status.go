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

//nolint:funlen
func (h *Handler) UpdateSettingStatus(
	ctx *gin.Context,
	resType base.ResourceType,
	scopeType base.SettingScopeType,
	opts ...UpdateSettingOption,
) {
	var auth *basedto.Auth
	var itemID string
	var err error

	options := &UpdateSettingOptions{}
	for _, o := range opts {
		o(options)
	}

	scope := &base.SettingScope{}
	switch scopeType {
	case base.SettingScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, resType, base.ActionTypeWrite, "itemID")
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
		r := basicauthdto.NewUpdateBasicAuthStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.BasicAuthUC.UpdateBasicAuthStatus(reqCtx, auth, r) }

	case base.ResourceTypeGithubApp:
		r := githubappdto.NewUpdateGithubAppStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.GithubAppUC.UpdateGithubAppStatus(reqCtx, auth, r) }

	case base.ResourceTypeAccessToken:
		r := accesstokendto.NewUpdateAccessTokenStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.AccessTokenUC.UpdateAccessTokenStatus(reqCtx, auth, r) }

	case base.ResourceTypeOAuth:
		r := oauthdto.NewUpdateOAuthStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.OAuthUC.UpdateOAuthStatus(reqCtx, auth, r) }

	case base.ResourceTypeRegistryAuth:
		r := registryauthdto.NewUpdateRegistryAuthStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RegistryAuthUC.UpdateRegistryAuthStatus(reqCtx, auth, r) }

	case base.ResourceTypeCloudStorage:
		r := cloudstoragedto.NewUpdateCloudStorageStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.CloudStorageUC.UpdateCloudStorageStatus(reqCtx, auth, r) }

	case base.ResourceTypeSSHKey:
		r := sshkeydto.NewUpdateSSHKeyStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSHKeyUC.UpdateSSHKeyStatus(reqCtx, auth, r) }

	case base.ResourceTypeSSLCert:
		r := sslcertdto.NewUpdateSSLCertStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SSLCertUC.UpdateSSLCertStatus(reqCtx, auth, r) }

	case base.ResourceTypeSchedJob:
		r := schedjobdto.NewUpdateSchedJobStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SchedJobUC.UpdateSchedJobStatus(reqCtx, auth, r) }

	case base.ResourceTypeHealthcheck:
		r := healthcheckdto.NewUpdateHealthcheckStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.HealthcheckUC.UpdateHealthcheckStatus(reqCtx, auth, r) }

	case base.ResourceTypeSecret:
		r := secretdto.NewUpdateSecretStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.SecretUC.UpdateSecretStatus(reqCtx, auth, r) }

	case base.ResourceTypeConfigFile:
		r := configfiledto.NewUpdateConfigFileStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.ConfigFileUC.UpdateConfigFileStatus(reqCtx, auth, r) }

	case base.ResourceTypeAPIKey:
		r := apikeydto.NewUpdateAPIKeyStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.APIKeyUC.UpdateAPIKeyStatus(reqCtx, auth, r) }

	case base.ResourceTypeIMService:
		r := imservicedto.NewUpdateIMServiceStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.IMServiceUC.UpdateIMServiceStatus(reqCtx, auth, r) }

	case base.ResourceTypeEmail:
		r := emaildto.NewUpdateEmailStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.EmailUC.UpdateEmailStatus(reqCtx, auth, r) }

	case base.ResourceTypeRepoWebhook:
		r := repowebhookdto.NewUpdateRepoWebhookStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.RepoWebhookUC.UpdateRepoWebhookStatus(reqCtx, auth, r) }

	case base.ResourceTypeNotification:
		r := notificationdto.NewUpdateNotificationStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.NotificationUC.UpdateNotificationStatus(reqCtx, auth, r) }

	case base.ResourceTypeFile:
		r := filedto.NewUpdateFileStatusReq()
		r.Scope, r.ID = scope, itemID
		req, ucFunc = r, func() (any, error) { return h.FileUC.UpdateFileStatus(reqCtx, auth, r) }

	default:
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

	ctx.JSON(http.StatusOK, resp)
}
