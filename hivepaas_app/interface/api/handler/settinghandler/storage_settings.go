package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

// GetStorageSettings Gets storage settings details
// @Summary Gets storage settings details
// @Description Gets storage settings details
// @Tags    settings
// @Produce json
// @Id      getSettingStorageSettings
// @Success 200 {object} storagesettingsdto.GetStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings [get]
func (h *Handler) GetStorageSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.ObjectScopeGlobal)
}

// UpdateStorageSettings Updates storage settings
// @Summary Updates storage settings
// @Description Updates storage settings
// @Tags    settings
// @Produce json
// @Id      updateSettingStorageSettings
// @Param   body body storagesettingsdto.UpdateStorageSettingsReq true "request data"
// @Success 200 {object} storagesettingsdto.UpdateStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings [put]
func (h *Handler) UpdateStorageSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.ObjectScopeGlobal)
}

// UpdateStorageSettingsStatus Updates storage settings status
// @Summary Updates storage settings status
// @Description Updates storage settings status
// @Tags    settings
// @Produce json
// @Id      updateSettingStorageSettingsStatus
// @Param   body body storagesettingsdto.UpdateStorageSettingsStatusReq true "request data"
// @Success 200 {object} storagesettingsdto.UpdateStorageSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings/status [put]
func (h *Handler) UpdateStorageSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeStorageSettings, base.ObjectScopeGlobal)
}

// DeleteStorageSettings Deletes storage settings setting
// @Summary Deletes storage settings setting
// @Description Deletes storage settings setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingStorageSettings
// @Success 200 {object} storagesettingsdto.DeleteStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings [delete]
func (h *Handler) DeleteStorageSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.ObjectScopeGlobal)
}
