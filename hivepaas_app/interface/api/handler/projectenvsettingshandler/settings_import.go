package projectenvsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

// ImportSettings Imports settings from global to a project
// @Summary Imports settings from global to a project
// @Description Imports settings from global to a project
// @Tags    project_env_settings
// @Produce json
// @Id      importSettingsToProjectEnv
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body projectsettingsdto.ImportSettingsToProjectReq true "request data"
// @Success 200 {object} projectsettingsdto.ImportSettingsToProjectResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/settings-import [post]
func (h *Handler) ImportSettings(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewImportSettingsToProjectReq()
	req.ProjectID = projectID
	req.ProjectEnv = projectEnvID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.ImportSettingsToProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
