package projecthandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/domainsettingsuc/domainsettingsdto"
)

// GetUniqueDomainSettings Gets domain settings details
// @Summary Gets domain settings details
// @Description Gets domain settings details
// @Tags    project_settings
// @Produce json
// @Id      getProjectDomainSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} domainsettingsdto.GetUniqueDomainSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/domain-settings [get]
func (h *Handler) GetUniqueDomainSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeDomainSettings, base.SettingScopeProject)
}

// UpdateUniqueDomainSettings Updates domain settings
// @Summary Updates domain settings
// @Description Updates domain settings
// @Tags    project_settings
// @Produce json
// @Id      updateProjectDomainSettings
// @Param   projectID path string true "project ID"
// @Param   body body domainsettingsdto.UpdateUniqueDomainSettingsReq true "request data"
// @Success 200 {object} domainsettingsdto.UpdateUniqueDomainSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/domain-settings [put]
func (h *Handler) UpdateUniqueDomainSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeDomainSettings, base.SettingScopeProject)
}

// UpdateUniqueDomainSettingsStatus Updates domain settings status
// @Summary Updates domain settings status
// @Description Updates domain settings status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectDomainSettingsStatus
// @Param   projectID path string true "project ID"
// @Param   body body domainsettingsdto.UpdateUniqueDomainSettingsStatusReq true "request data"
// @Success 200 {object} domainsettingsdto.UpdateUniqueDomainSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/domain-settings/status [put]
func (h *Handler) UpdateUniqueDomainSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeDomainSettings, base.SettingScopeProject)
}

// DeleteUniqueDomainSettings Deletes domain settings setting
// @Summary Deletes domain settings setting
// @Description Deletes domain settings setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectDomainSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} domainsettingsdto.DeleteUniqueDomainSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/domain-settings [delete]
func (h *Handler) DeleteUniqueDomainSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeDomainSettings, base.SettingScopeProject)
}
