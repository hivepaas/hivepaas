package appsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

// ListSecret Lists app secrets
// @Summary Lists app secrets
// @Description Lists app secrets
// @Tags    app_settings
// @Produce json
// @Id      listAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} secretdto.ListSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets [get]
func (h *Handler) ListSecret(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// GetSecret Get an app secret details
// @Summary Get an app secret details
// @Description Get an app secret details
// @Tags    app_settings
// @Produce json
// @Id      getAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} secretdto.GetSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets/{itemID} [get]
func (h *Handler) GetSecret(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// GetSecretDownloadToken Gets secret download token
// @Summary Gets secret download token
// @Description Gets secret download token
// @Tags    app_settings
// @Produce json
// @Id      getAppSecretDownloadToken
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} secretdto.GetDownloadTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets/{itemID}/download-token [get]
func (h *Handler) GetSecretDownloadToken(ctx *gin.Context) {
	h.GetDownloadToken(ctx, base.ResourceTypeSecret, base.SettingScopeApp, "", 0)
}

// DownloadSecret Download a secret as a file
// @Summary Download a secret as a file
// @Description Download a secret as a file
// @Tags    app_settings
// @Produce json
// @Id      downloadAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} secretdto.DownloadSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets/{itemID}/download [get]
func (h *Handler) DownloadSecret(ctx *gin.Context) {
	h.Download(ctx, base.ResourceTypeSecret, base.SettingScopeApp, "")
}

// CreateSecret Creates an app secret
// @Summary Creates an app secret
// @Description Creates an app secret
// @Tags    app_settings
// @Produce json
// @Id      createAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body secretdto.CreateSecretReq true "request data"
// @Success 201 {object} secretdto.CreateSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets [post]
func (h *Handler) CreateSecret(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// UpdateSecret Updates an app secret
// @Summary Updates an app secret
// @Description Updates an app secret
// @Tags    app_settings
// @Produce json
// @Id      updateAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body secretdto.UpdateSecretReq true "request data"
// @Success 200 {object} secretdto.UpdateSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets/{itemID} [put]
func (h *Handler) UpdateSecret(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// UpdateSecretStatus Updates app secret status
// @Summary Updates app secret status
// @Description Updates app secret status
// @Tags    app_settings
// @Produce json
// @Id      updateAppSecretStatus
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body secretdto.UpdateSecretStatusReq true "request data"
// @Success 200 {object} secretdto.UpdateSecretStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets/{itemID}/status [put]
func (h *Handler) UpdateSecretStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// DeleteSecret Deletes an app secret
// @Summary Deletes an app secret
// @Description Deletes an app secret
// @Tags    app_settings
// @Produce json
// @Id      deleteAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} secretdto.DeleteSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets/{itemID} [delete]
func (h *Handler) DeleteSecret(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}
