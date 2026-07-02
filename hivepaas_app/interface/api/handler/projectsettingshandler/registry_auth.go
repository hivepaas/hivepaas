package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
)

// ListRegistryAuth Lists registry auth settings
// @Summary Lists registry auth settings
// @Description Lists registry auth settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} registryauthdto.ListRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/registry-auth [get]
func (h *Handler) ListRegistryAuth(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProject)
}

// GetRegistryAuth Gets registry auth setting details
// @Summary Gets registry auth setting details
// @Description Gets registry auth setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} registryauthdto.GetRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/registry-auth/{itemID} [get]
func (h *Handler) GetRegistryAuth(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProject)
}

// CreateRegistryAuth Creates a new registry auth setting
// @Summary Creates a new registry auth setting
// @Description Creates a new registry auth setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   body body registryauthdto.CreateRegistryAuthReq true "request data"
// @Success 201 {object} registryauthdto.CreateRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/registry-auth [post]
func (h *Handler) CreateRegistryAuth(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProject)
}

// UpdateRegistryAuth Updates registry auth
// @Summary Updates registry auth
// @Description Updates registry auth
// @Tags    project_settings
// @Produce json
// @Id      updateProjectRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body registryauthdto.UpdateRegistryAuthReq true "request data"
// @Success 200 {object} registryauthdto.UpdateRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/registry-auth/{itemID} [put]
func (h *Handler) UpdateRegistryAuth(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProject)
}

// UpdateRegistryAuthStatus Updates registry auth status
// @Summary Updates registry auth status
// @Description Updates registry auth status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectRegistryAuthStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body registryauthdto.UpdateRegistryAuthStatusReq true "request data"
// @Success 200 {object} registryauthdto.UpdateRegistryAuthStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/registry-auth/{itemID}/status [put]
func (h *Handler) UpdateRegistryAuthStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProject)
}

// DeleteRegistryAuth Deletes registry auth setting
// @Summary Deletes registry auth setting
// @Description Deletes registry auth setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectRegistryAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} registryauthdto.DeleteRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/registry-auth/{itemID} [delete]
func (h *Handler) DeleteRegistryAuth(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeProject)
}
