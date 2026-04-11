package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/secretuc/secretdto"
)

// ListProjectSecrets Lists project secrets
// @Summary Lists project secrets
// @Description Lists project secrets
// @Tags    projects
// @Produce json
// @Id      listProjectSecret
// @Param   projectID path string true "project ID"
// @Param   type query string false "`type=<setting type>`"
// @Success 200 {object} secretdto.ListSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/secrets [get]
func (h *Handler) ListProjectSecrets(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSecret, base.SettingScopeProject)
}

// CreateProjectSecret Creates a project secret
// @Summary Creates a project secret
// @Description Creates a project secret
// @Tags    projects
// @Produce json
// @Id      createProjectSecret
// @Param   projectID path string true "project ID"
// @Param   body body secretdto.CreateSecretReq true "request data"
// @Success 201 {object} secretdto.CreateSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/secrets [post]
func (h *Handler) CreateProjectSecret(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSecret, base.SettingScopeProject)
}

// UpdateProjectSecret Updates a project secret
// @Summary Updates a project secret
// @Description Updates a project secret
// @Tags    projects
// @Produce json
// @Id      updateProjectSecret
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body secretdto.UpdateSecretReq true "request data"
// @Success 200 {object} secretdto.UpdateSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/secrets/{itemID} [put]
func (h *Handler) UpdateProjectSecret(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSecret, base.SettingScopeProject)
}

// DeleteProjectSecret Deletes a project secret
// @Summary Deletes a project secret
// @Description Deletes a project secret
// @Tags    projects
// @Produce json
// @Id      deleteProjectSecret
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} secretdto.DeleteSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/secrets/{itemID} [delete]
func (h *Handler) DeleteProjectSecret(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSecret, base.SettingScopeProject)
}
