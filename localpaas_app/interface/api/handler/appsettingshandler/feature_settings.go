package appsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/appfeaturesettingsuc/appfeaturesettingsdto"
)

// GetAppFeatureSettings Gets app feature settings
// @Summary Gets app feature settings
// @Description Gets app feature settings
// @Tags    app_settings
// @Produce json
// @Id      getAppFeatureSettings
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} appfeaturesettingsdto.GetAppFeatureSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/feature-settings [get]
func (h *Handler) GetAppFeatureSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeAppFeatures, base.ObjectScopeApp)
}

// UpdateAppFeatureSettings Updates app feature settings
// @Summary Updates app feature settings
// @Description Updates app feature settings
// @Tags    app_settings
// @Produce json
// @Id      updateAppFeatureSettings
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body appfeaturesettingsdto.UpdateAppFeatureSettingsReq true "request data"
// @Success 200 {object} appfeaturesettingsdto.UpdateAppFeatureSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/feature-settings [put]
func (h *Handler) UpdateAppFeatureSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeAppFeatures, base.ObjectScopeApp)
}
