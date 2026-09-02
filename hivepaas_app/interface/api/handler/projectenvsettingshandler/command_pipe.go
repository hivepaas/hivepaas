package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
)

// ListCommandPipe Lists command pipe settings
// @Summary Lists command pipe settings
// @Description Lists command pipe settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvCommandPipe
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} commandpipedto.ListCommandPipeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-pipes [get]
func (h *Handler) ListCommandPipe(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProjectEnv)
}

// GetCommandPipe Gets command pipe setting details
// @Summary Gets command pipe setting details
// @Description Gets command pipe setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvCommandPipe
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandpipedto.GetCommandPipeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-pipes/{itemID} [get]
func (h *Handler) GetCommandPipe(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProjectEnv)
}

// CreateCommandPipe Creates a new command pipe setting
// @Summary Creates a new command pipe setting
// @Description Creates a new command pipe setting
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvCommandPipe
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body commandpipedto.CreateCommandPipeReq true "request data"
// @Success 201 {object} commandpipedto.CreateCommandPipeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-pipes [post]
func (h *Handler) CreateCommandPipe(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProjectEnv)
}

// CreateCommandPipeFromTemplate Creates a command pipe setting from template
// @Summary Creates a command pipe setting from template
// @Description Creates a command pipe setting from template
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvCommandPipeFromTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body commandpipedto.CreateCommandPipeFromTemplateReq true "request data"
// @Success 201 {object} commandpipedto.CreateCommandPipeFromTemplateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-pipes/from-template [post]
func (h *Handler) CreateCommandPipeFromTemplate(ctx *gin.Context) {
	h.CommandPipeCreateFromTemplate(ctx, base.ObjectScopeProjectEnv)
}

// UpdateCommandPipe Updates command pipe
// @Summary Updates command pipe
// @Description Updates command pipe
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvCommandPipe
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body commandpipedto.UpdateCommandPipeReq true "request data"
// @Success 200 {object} commandpipedto.UpdateCommandPipeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-pipes/{itemID} [put]
func (h *Handler) UpdateCommandPipe(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProjectEnv)
}

// UpdateCommandPipeStatus Updates command pipe status
// @Summary Updates command pipe status
// @Description Updates command pipe status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvCommandPipeStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body commandpipedto.UpdateCommandPipeStatusReq true "request data"
// @Success 200 {object} commandpipedto.UpdateCommandPipeStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-pipes/{itemID}/status [put]
func (h *Handler) UpdateCommandPipeStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProjectEnv)
}

// DeleteCommandPipe Deletes command pipe setting
// @Summary Deletes command pipe setting
// @Description Deletes command pipe setting
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvCommandPipe
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandpipedto.DeleteCommandPipeResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/command-pipes/{itemID} [delete]
func (h *Handler) DeleteCommandPipe(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProjectEnv)
}
