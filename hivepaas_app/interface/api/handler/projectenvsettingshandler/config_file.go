package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
)

// ListConfigFile Lists project config files
// @Summary Lists project config files
// @Description Lists project config files
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvConfigFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   type query string false "`type=<setting type>`"
// @Success 200 {object} configfiledto.ListConfigFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/config-files [get]
func (h *Handler) ListConfigFile(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeProjectEnv)
}

// GetConfigFile Gets config file details
// @Summary Gets config file details
// @Description Gets config file details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvConfigFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} configfiledto.GetConfigFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/config-files/{itemID} [get]
func (h *Handler) GetConfigFile(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeProjectEnv)
}

// CreateConfigFile Creates a project config file
// @Summary Creates a project config file
// @Description Creates a project config file
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvConfigFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body configfiledto.CreateConfigFileReq true "request data"
// @Success 201 {object} configfiledto.CreateConfigFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/config-files [post]
func (h *Handler) CreateConfigFile(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeProjectEnv)
}

// UpdateConfigFile Updates a project config file
// @Summary Updates a project config file
// @Description Updates a project config file
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvConfigFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body configfiledto.UpdateConfigFileReq true "request data"
// @Success 200 {object} configfiledto.UpdateConfigFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/config-files/{itemID} [put]
func (h *Handler) UpdateConfigFile(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeProjectEnv)
}

// UpdateConfigFileStatus Updates project config file status
// @Summary Updates project config file status
// @Description Updates project config file status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvConfigFileStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body configfiledto.UpdateConfigFileStatusReq true "request data"
// @Success 200 {object} configfiledto.UpdateConfigFileStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/config-files/{itemID}/status [put]
func (h *Handler) UpdateConfigFileStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeConfigFile, base.ObjectScopeProjectEnv)
}

// DeleteConfigFile Deletes a project config file
// @Summary Deletes a project config file
// @Description Deletes a project config file
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvConfigFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} configfiledto.DeleteConfigFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/config-files/{itemID} [delete]
func (h *Handler) DeleteConfigFile(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeProjectEnv)
}
