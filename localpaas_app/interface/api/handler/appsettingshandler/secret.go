package appsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

// ListAppSecret Lists app secrets
// @Summary Lists app secrets
// @Description Lists app secrets
// @Tags    apps
// @Produce json
// @Id      listAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} secretdto.ListSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets [get]
func (h *Handler) ListAppSecret(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// CreateAppSecret Creates an app secret
// @Summary Creates an app secret
// @Description Creates an app secret
// @Tags    apps
// @Produce json
// @Id      createAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body secretdto.CreateSecretReq true "request data"
// @Success 201 {object} secretdto.CreateSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets [post]
func (h *Handler) CreateAppSecret(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// UpdateAppSecret Updates an app secret
// @Summary Updates an app secret
// @Description Updates an app secret
// @Tags    apps
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
func (h *Handler) UpdateAppSecret(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// UpdateAppSecretStatus Updates app secret status
// @Summary Updates app secret status
// @Description Updates app secret status
// @Tags    apps
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
func (h *Handler) UpdateAppSecretStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}

// DeleteAppSecret Deletes an app secret
// @Summary Deletes an app secret
// @Description Deletes an app secret
// @Tags    apps
// @Produce json
// @Id      deleteAppSecret
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} secretdto.DeleteSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/secrets/{itemID} [delete]
func (h *Handler) DeleteAppSecret(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSecret, base.SettingScopeApp)
}
