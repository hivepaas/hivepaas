package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

// ListCommandTemplate Lists command template settings
// @Summary Lists command template settings
// @Description Lists command template settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} commandtemplatedto.ListCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-templates [get]
func (h *Handler) ListCommandTemplate(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProject)
}

// GetCommandTemplate Gets command template setting details
// @Summary Gets command template setting details
// @Description Gets command template setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandtemplatedto.GetCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-templates/{itemID} [get]
func (h *Handler) GetCommandTemplate(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProject)
}

// CreateCommandTemplate Creates a new command template setting
// @Summary Creates a new command template setting
// @Description Creates a new command template setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   body body commandtemplatedto.CreateCommandTemplateReq true "request data"
// @Success 201 {object} commandtemplatedto.CreateCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-templates [post]
func (h *Handler) CreateCommandTemplate(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProject)
}

// UpdateCommandTemplate Updates command template
// @Summary Updates command template
// @Description Updates command template
// @Tags    project_settings
// @Produce json
// @Id      updateProjectCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body commandtemplatedto.UpdateCommandTemplateReq true "request data"
// @Success 200 {object} commandtemplatedto.UpdateCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-templates/{itemID} [put]
func (h *Handler) UpdateCommandTemplate(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProject)
}

// UpdateCommandTemplateStatus Updates command template status
// @Summary Updates command template status
// @Description Updates command template status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectCommandTemplateStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body commandtemplatedto.UpdateCommandTemplateStatusReq true "request data"
// @Success 200 {object} commandtemplatedto.UpdateCommandTemplateStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-templates/{itemID}/status [put]
func (h *Handler) UpdateCommandTemplateStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProject)
}

// DeleteCommandTemplate Deletes command template setting
// @Summary Deletes command template setting
// @Description Deletes command template setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandtemplatedto.DeleteCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-templates/{itemID} [delete]
func (h *Handler) DeleteCommandTemplate(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProject)
}
