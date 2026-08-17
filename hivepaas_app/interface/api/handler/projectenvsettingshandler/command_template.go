package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandtemplateuc/commandtemplatedto"
)

// ListCommandTemplate Lists command template settings
// @Summary Lists command template settings
// @Description Lists command template settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} commandtemplatedto.ListCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-templates [get]
func (h *Handler) ListCommandTemplate(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProjectEnv)
}

// GetCommandTemplate Gets command template setting details
// @Summary Gets command template setting details
// @Description Gets command template setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandtemplatedto.GetCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-templates/{itemID} [get]
func (h *Handler) GetCommandTemplate(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProjectEnv)
}

// CreateCommandTemplate Creates a new command template setting
// @Summary Creates a new command template setting
// @Description Creates a new command template setting
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body commandtemplatedto.CreateCommandTemplateReq true "request data"
// @Success 201 {object} commandtemplatedto.CreateCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-templates [post]
func (h *Handler) CreateCommandTemplate(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProjectEnv)
}

// CreateCommandTemplateFromTemplate Creates a new command template setting from a template
// @Summary Creates a new command template setting from a template
// @Description Creates a new command template setting from a template
// @Tags    project_settings
// @Produce json
// @Id      createProjectCommandTemplateFromTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body commandtemplatedto.CreateCommandTemplateFromTemplateReq true "request data"
// @Success 201 {object} commandtemplatedto.CreateCommandTemplateFromTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-templates/from-template [post]
func (h *Handler) CreateCommandTemplateFromTemplate(ctx *gin.Context) {
	h.Handler.CreateCommandTemplateFromTemplate(ctx, base.ObjectScopeProject)
}

// UpdateCommandTemplate Updates command template
// @Summary Updates command template
// @Description Updates command template
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body commandtemplatedto.UpdateCommandTemplateReq true "request data"
// @Success 200 {object} commandtemplatedto.UpdateCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-templates/{itemID} [put]
func (h *Handler) UpdateCommandTemplate(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProjectEnv)
}

// UpdateCommandTemplateStatus Updates command template status
// @Summary Updates command template status
// @Description Updates command template status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvCommandTemplateStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body commandtemplatedto.UpdateCommandTemplateStatusReq true "request data"
// @Success 200 {object} commandtemplatedto.UpdateCommandTemplateStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-templates/{itemID}/status [put]
func (h *Handler) UpdateCommandTemplateStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProjectEnv)
}

// DeleteCommandTemplate Deletes command template setting
// @Summary Deletes command template setting
// @Description Deletes command template setting
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvCommandTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandtemplatedto.DeleteCommandTemplateResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-templates/{itemID} [delete]
func (h *Handler) DeleteCommandTemplate(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCommandTemplate, base.ObjectScopeProjectEnv)
}
