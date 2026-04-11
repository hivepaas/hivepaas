package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

// GetUniqueImageBuildSettings Gets image build setting details
// @Summary Gets image build setting details
// @Description Gets image build setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectImageBuildSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} imagebuildsettingsdto.GetUniqueImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings [get]
func (h *Handler) GetUniqueImageBuildSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.SettingScopeProject)
}

// UpdateUniqueImageBuildSettings Updates image build settings
// @Summary Updates image build settings
// @Description Updates image build settings
// @Tags    project_settings
// @Produce json
// @Id      updateProjectImageBuildSettings
// @Param   projectID path string true "project ID"
// @Param   body body imagebuildsettingsdto.UpdateUniqueImageBuildSettingsReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateUniqueImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings [put]
func (h *Handler) UpdateUniqueImageBuildSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.SettingScopeProject)
}

// UpdateUniqueImageBuildSettingsStatus Updates image build status
// @Summary Updates image build status
// @Description Updates image build status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectImageBuildSettingsStatus
// @Param   projectID path string true "project ID"
// @Param   body body imagebuildsettingsdto.UpdateUniqueImageBuildSettingsStatusReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateUniqueImageBuildSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings/status [put]
func (h *Handler) UpdateUniqueImageBuildSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeImageBuildSettings, base.SettingScopeProject)
}

// DeleteUniqueImageBuildSettings Deletes image build settings
// @Summary Deletes image build settings
// @Description Deletes image build settings
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectImageBuildSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} imagebuildsettingsdto.DeleteUniqueImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/image-build-settings [delete]
func (h *Handler) DeleteUniqueImageBuildSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.SettingScopeProject)
}
