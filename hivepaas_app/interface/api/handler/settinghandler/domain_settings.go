package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

// GetDomainSettings Gets domain settings details
// @Summary Gets domain settings details
// @Description Gets domain settings details
// @Tags    settings
// @Produce json
// @Id      getSettingDomainSettings
// @Success 200 {object} domainsettingsdto.GetDomainSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/domain-settings [get]
func (h *Handler) GetDomainSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeDomainSettings, base.ObjectScopeGlobal)
}

// UpdateDomainSettings Updates domain settings
// @Summary Updates domain settings
// @Description Updates domain settings
// @Tags    settings
// @Produce json
// @Id      updateSettingDomainSettings
// @Param   body body domainsettingsdto.UpdateDomainSettingsReq true "request data"
// @Success 200 {object} domainsettingsdto.UpdateDomainSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/domain-settings [put]
func (h *Handler) UpdateDomainSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeDomainSettings, base.ObjectScopeGlobal)
}

// UpdateDomainSettingsStatus Updates domain settings status
// @Summary Updates domain settings status
// @Description Updates domain settings status
// @Tags    settings
// @Produce json
// @Id      updateSettingDomainSettingsStatus
// @Param   body body domainsettingsdto.UpdateDomainSettingsStatusReq true "request data"
// @Success 200 {object} domainsettingsdto.UpdateDomainSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/domain-settings/status [put]
func (h *Handler) UpdateDomainSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeDomainSettings, base.ObjectScopeGlobal)
}

// DeleteDomainSettings Deletes domain settings setting
// @Summary Deletes domain settings setting
// @Description Deletes domain settings setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingDomainSettings
// @Success 200 {object} domainsettingsdto.DeleteDomainSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/domain-settings [delete]
func (h *Handler) DeleteDomainSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeDomainSettings, base.ObjectScopeGlobal)
}
