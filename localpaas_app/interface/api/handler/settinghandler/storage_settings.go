package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

// GetUniqueStorageSettings Gets storage settings details
// @Summary Gets storage settings details
// @Description Gets storage settings details
// @Tags    settings
// @Produce json
// @Id      getSettingStorageSettings
// @Success 200 {object} storagesettingsdto.GetUniqueStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings [get]
func (h *Handler) GetUniqueStorageSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.SettingScopeGlobal)
}

// UpdateUniqueStorageSettings Updates storage settings
// @Summary Updates storage settings
// @Description Updates storage settings
// @Tags    settings
// @Produce json
// @Id      updateSettingStorageSettings
// @Param   body body storagesettingsdto.UpdateUniqueStorageSettingsReq true "request data"
// @Success 200 {object} storagesettingsdto.UpdateUniqueStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings [put]
func (h *Handler) UpdateUniqueStorageSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.SettingScopeGlobal)
}

// UpdateUniqueStorageSettingsStatus Updates storage settings status
// @Summary Updates storage settings status
// @Description Updates storage settings status
// @Tags    settings
// @Produce json
// @Id      updateSettingStorageSettingsStatus
// @Param   body body storagesettingsdto.UpdateUniqueStorageSettingsStatusReq true "request data"
// @Success 200 {object} storagesettingsdto.UpdateUniqueStorageSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings/status [put]
func (h *Handler) UpdateUniqueStorageSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeStorageSettings, base.SettingScopeGlobal)
}

// DeleteUniqueStorageSettings Deletes storage settings setting
// @Summary Deletes storage settings setting
// @Description Deletes storage settings setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingStorageSettings
// @Success 200 {object} storagesettingsdto.DeleteUniqueStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/storage-settings [delete]
func (h *Handler) DeleteUniqueStorageSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.SettingScopeGlobal)
}
