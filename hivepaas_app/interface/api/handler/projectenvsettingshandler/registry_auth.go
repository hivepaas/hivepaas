package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
)

// ListRegistryAuth Lists registry auth settings
// @Summary Lists registry auth settings
// @Description Lists registry auth settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} registryauthdto.ListRegistryAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/registry-auth [get]
func (h *Handler) ListRegistryAuth(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProjectEnv)
}

// GetRegistryAuth Gets registry auth setting details
// @Summary Gets registry auth setting details
// @Description Gets registry auth setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} registryauthdto.GetRegistryAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/registry-auth/{itemID} [get]
func (h *Handler) GetRegistryAuth(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProjectEnv)
}

// CreateRegistryAuth Creates a new registry auth setting
// @Summary Creates a new registry auth setting
// @Description Creates a new registry auth setting
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body registryauthdto.CreateRegistryAuthReq true "request data"
// @Success 201 {object} registryauthdto.CreateRegistryAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/registry-auth [post]
func (h *Handler) CreateRegistryAuth(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProjectEnv)
}

// UpdateRegistryAuth Updates registry auth
// @Summary Updates registry auth
// @Description Updates registry auth
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body registryauthdto.UpdateRegistryAuthReq true "request data"
// @Success 200 {object} registryauthdto.UpdateRegistryAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/registry-auth/{itemID} [put]
func (h *Handler) UpdateRegistryAuth(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProjectEnv)
}

// UpdateRegistryAuthStatus Updates registry auth status
// @Summary Updates registry auth status
// @Description Updates registry auth status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvRegistryAuthStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body registryauthdto.UpdateRegistryAuthStatusReq true "request data"
// @Success 200 {object} registryauthdto.UpdateRegistryAuthStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/registry-auth/{itemID}/status [put]
func (h *Handler) UpdateRegistryAuthStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProjectEnv)
}

// DeleteRegistryAuth Deletes registry auth setting
// @Summary Deletes registry auth setting
// @Description Deletes registry auth setting
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} registryauthdto.DeleteRegistryAuthResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/registry-auth/{itemID} [delete]
func (h *Handler) DeleteRegistryAuth(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProjectEnv)
}
