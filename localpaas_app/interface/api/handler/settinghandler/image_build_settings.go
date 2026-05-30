package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

// GetImageBuildSettings Gets image build settings
// @Summary Gets image build settings
// @Description Gets image build settings
// @Tags    settings
// @Produce json
// @Id      getSettingImageBuildSettings
// @Success 200 {object} imagebuildsettingsdto.GetImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings [get]
func (h *Handler) GetImageBuildSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}

// UpdateImageBuildSettings Updates image build settings
// @Summary Updates image build settings
// @Description Updates image build settings
// @Tags    settings
// @Produce json
// @Id      updateSettingImageBuildSettings
// @Param   body body imagebuildsettingsdto.UpdateImageBuildSettingsReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings [put]
func (h *Handler) UpdateImageBuildSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}

// UpdateImageBuildSettingsStatus Updates image build settings status
// @Summary Updates image build settings status
// @Description Updates image build settings status
// @Tags    settings
// @Produce json
// @Id      updateSettingImageBuildSettingsStatus
// @Param   body body imagebuildsettingsdto.UpdateImageBuildSettingsStatusReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateImageBuildSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings/status [put]
func (h *Handler) UpdateImageBuildSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}

// DeleteImageBuildSettings Deletes image build settings
// @Summary Deletes image build settings
// @Description Deletes image build settings
// @Tags    settings
// @Produce json
// @Id      deleteSettingImageBuildSettings
// @Success 200 {object} imagebuildsettingsdto.DeleteImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings [delete]
func (h *Handler) DeleteImageBuildSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}
