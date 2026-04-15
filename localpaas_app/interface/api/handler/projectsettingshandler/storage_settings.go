package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/storagesettingsuc/storagesettingsdto"
)

// GetUniqueStorageSettings Gets storage settings details
// @Summary Gets storage settings details
// @Description Gets storage settings details
// @Tags    project_settings
// @Produce json
// @Id      getProjectStorageSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} storagesettingsdto.GetUniqueStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/storage-settings [get]
func (h *Handler) GetUniqueStorageSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.SettingScopeProject)
}

// UpdateUniqueStorageSettings Updates storage settings
// @Summary Updates storage settings
// @Description Updates storage settings
// @Tags    project_settings
// @Produce json
// @Id      updateProjectStorageSettings
// @Param   projectID path string true "project ID"
// @Param   body body storagesettingsdto.UpdateUniqueStorageSettingsReq true "request data"
// @Success 200 {object} storagesettingsdto.UpdateUniqueStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/storage-settings [put]
func (h *Handler) UpdateUniqueStorageSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.SettingScopeProject)
}

// UpdateUniqueStorageSettingsStatus Updates storage settings status
// @Summary Updates storage settings status
// @Description Updates storage settings status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectStorageSettingsStatus
// @Param   projectID path string true "project ID"
// @Param   body body storagesettingsdto.UpdateUniqueStorageSettingsStatusReq true "request data"
// @Success 200 {object} storagesettingsdto.UpdateUniqueStorageSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/storage-settings/status [put]
func (h *Handler) UpdateUniqueStorageSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeStorageSettings, base.SettingScopeProject)
}

// DeleteUniqueStorageSettings Deletes storage settings setting
// @Summary Deletes storage settings setting
// @Description Deletes storage settings setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectStorageSettings
// @Param   projectID path string true "project ID"
// @Success 200 {object} storagesettingsdto.DeleteUniqueStorageSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/storage-settings [delete]
func (h *Handler) DeleteUniqueStorageSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeStorageSettings, base.SettingScopeProject)
}
