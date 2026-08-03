package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/commandpipeuc/commandpipedto"
)

// ListCommandPipe Lists command pipe settings
// @Summary Lists command pipe settings
// @Description Lists command pipe settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectCommandPipe
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} commandpipedto.ListCommandPipeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-pipes [get]
func (h *Handler) ListCommandPipe(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProject)
}

// GetCommandPipe Gets command pipe setting details
// @Summary Gets command pipe setting details
// @Description Gets command pipe setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectCommandPipe
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandpipedto.GetCommandPipeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-pipes/{itemID} [get]
func (h *Handler) GetCommandPipe(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProject)
}

// CreateCommandPipe Creates a new command pipe setting
// @Summary Creates a new command pipe setting
// @Description Creates a new command pipe setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectCommandPipe
// @Param   projectID path string true "project ID"
// @Param   body body commandpipedto.CreateCommandPipeReq true "request data"
// @Success 201 {object} commandpipedto.CreateCommandPipeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-pipes [post]
func (h *Handler) CreateCommandPipe(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProject)
}

// UpdateCommandPipe Updates command pipe
// @Summary Updates command pipe
// @Description Updates command pipe
// @Tags    project_settings
// @Produce json
// @Id      updateProjectCommandPipe
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body commandpipedto.UpdateCommandPipeReq true "request data"
// @Success 200 {object} commandpipedto.UpdateCommandPipeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-pipes/{itemID} [put]
func (h *Handler) UpdateCommandPipe(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProject)
}

// UpdateCommandPipeStatus Updates command pipe status
// @Summary Updates command pipe status
// @Description Updates command pipe status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectCommandPipeStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body commandpipedto.UpdateCommandPipeStatusReq true "request data"
// @Success 200 {object} commandpipedto.UpdateCommandPipeStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-pipes/{itemID}/status [put]
func (h *Handler) UpdateCommandPipeStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProject)
}

// DeleteCommandPipe Deletes command pipe setting
// @Summary Deletes command pipe setting
// @Description Deletes command pipe setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectCommandPipe
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} commandpipedto.DeleteCommandPipeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/command-pipes/{itemID} [delete]
func (h *Handler) DeleteCommandPipe(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCommandPipe, base.ObjectScopeProject)
}
