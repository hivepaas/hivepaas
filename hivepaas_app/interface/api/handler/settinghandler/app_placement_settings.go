package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/appplacementsettingsuc/appplacementsettingsdto"
)

// GetAppPlacementSettings Gets app placement settings
// @Summary Gets app placement settings
// @Description Gets app placement settings
// @Tags    settings
// @Produce json
// @Id      getSettingAppPlacementSettings
// @Success 200 {object} appplacementsettingsdto.GetAppPlacementSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/app-placement-settings [get]
func (h *Handler) GetAppPlacementSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeAppPlacement, base.ObjectScopeGlobal)
}

// UpdateAppPlacementSettings Updates app placement settings
// @Summary Updates app placement settings
// @Description Updates app placement settings
// @Tags    settings
// @Produce json
// @Id      updateSettingAppPlacementSettings
// @Param   body body appplacementsettingsdto.UpdateAppPlacementSettingsReq true "request data"
// @Success 200 {object} appplacementsettingsdto.UpdateAppPlacementSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/app-placement-settings [put]
func (h *Handler) UpdateAppPlacementSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeAppPlacement, base.ObjectScopeGlobal)
}

// UpdateAppPlacementSettingsStatus Updates app placement settings status
// @Summary Updates app placement settings status
// @Description Updates app placement settings status
// @Tags    settings
// @Produce json
// @Id      updateSettingAppPlacementSettingsStatus
// @Param   body body appplacementsettingsdto.UpdateAppPlacementSettingsStatusReq true "request data"
// @Success 200 {object} appplacementsettingsdto.UpdateAppPlacementSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/app-placement-settings/status [put]
func (h *Handler) UpdateAppPlacementSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeAppPlacement, base.ObjectScopeGlobal)
}

// DeleteAppPlacementSettings Deletes app placement settings
// @Summary Deletes app placement settings
// @Description Deletes app placement settings
// @Tags    settings
// @Produce json
// @Id      deleteSettingAppPlacementSettings
// @Success 200 {object} appplacementsettingsdto.DeleteAppPlacementSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/app-placement-settings [delete]
func (h *Handler) DeleteAppPlacementSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeAppPlacement, base.ObjectScopeGlobal)
}
