package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
)

// ListBasicAuth Lists basic auth settings
// @Summary Lists basic auth settings
// @Description Lists basic auth settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectBasicAuth
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} basicauthdto.ListBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/basic-auth [get]
func (h *Handler) ListBasicAuth(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeBasicAuth, base.SettingScopeProject)
}

// GetBasicAuth Gets basic auth setting details
// @Summary Gets basic auth setting details
// @Description Gets basic auth setting details
// @Tags    project_settings
// @Produce json
// @Id      getProjectBasicAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} basicauthdto.GetBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/basic-auth/{itemID} [get]
func (h *Handler) GetBasicAuth(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeBasicAuth, base.SettingScopeProject)
}

// CreateBasicAuth Creates a new basic auth setting
// @Summary Creates a new basic auth setting
// @Description Creates a new basic auth setting
// @Tags    project_settings
// @Produce json
// @Id      createProjectBasicAuth
// @Param   projectID path string true "project ID"
// @Param   body body basicauthdto.CreateBasicAuthReq true "request data"
// @Success 201 {object} basicauthdto.CreateBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/basic-auth [post]
func (h *Handler) CreateBasicAuth(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeBasicAuth, base.SettingScopeProject)
}

// UpdateBasicAuth Updates basic auth
// @Summary Updates basic auth
// @Description Updates basic auth
// @Tags    project_settings
// @Produce json
// @Id      updateProjectBasicAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body basicauthdto.UpdateBasicAuthReq true "request data"
// @Success 200 {object} basicauthdto.UpdateBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/basic-auth/{itemID} [put]
func (h *Handler) UpdateBasicAuth(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeBasicAuth, base.SettingScopeProject)
}

// UpdateBasicAuthStatus Updates basic auth status
// @Summary Updates basic auth status
// @Description Updates basic auth status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectBasicAuthStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body basicauthdto.UpdateBasicAuthStatusReq true "request data"
// @Success 200 {object} basicauthdto.UpdateBasicAuthStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/basic-auth/{itemID}/status [put]
func (h *Handler) UpdateBasicAuthStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeBasicAuth, base.SettingScopeProject)
}

// DeleteBasicAuth Deletes basic auth setting
// @Summary Deletes basic auth setting
// @Description Deletes basic auth setting
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectBasicAuth
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} basicauthdto.DeleteBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/basic-auth/{itemID} [delete]
func (h *Handler) DeleteBasicAuth(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeBasicAuth, base.SettingScopeProject)
}
