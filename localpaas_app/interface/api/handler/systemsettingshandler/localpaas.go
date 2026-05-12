package systemsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/localpaassettingsuc/localpaassettingsdto"
)

// GetLocalPaaSSettings Gets LocalPaaS settings
// @Summary Gets LocalPaaS settings
// @Description Gets LocalPaaS settings
// @Tags    system_settings
// @Produce json
// @Id      getSystemLocalPaaSSettings
// @Success 200 {object} localpaassettingsdto.GetLocalPaaSSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/localpaas [get]
func (h *Handler) GetLocalPaaSSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeLocalPaaSSettings, base.SettingScopeGlobal)
}

// UpdateLocalPaaSSettings Updates LocalPaaS settings
// @Summary Updates LocalPaaS settings
// @Description Updates LocalPaaS settings
// @Tags    system_settings
// @Produce json
// @Id      updateSystemLocalPaaSSettings
// @Param   body body localpaassettingsdto.UpdateLocalPaaSSettingsReq true "request data"
// @Success 200 {object} localpaassettingsdto.UpdateLocalPaaSSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/localpaas [put]
func (h *Handler) UpdateLocalPaaSSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeLocalPaaSSettings, base.SettingScopeGlobal)
}
