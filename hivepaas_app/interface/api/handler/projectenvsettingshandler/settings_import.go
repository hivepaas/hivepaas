package projectenvsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc/projectenvsettingsdto"
)

// ImportSettings Imports settings from global to a project
// @Summary Imports settings from global to a project
// @Description Imports settings from global to a project
// @Tags    project_env_settings
// @Produce json
// @Id      importSettingsToProjectEnv
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body projectenvsettingsdto.ImportSettingsToProjectEnvReq true "request data"
// @Success 200 {object} projectenvsettingsdto.ImportSettingsToProjectEnvResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/settings-import [post]
func (h *Handler) ImportSettings(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectenvsettingsdto.NewImportSettingsToProjectEnvReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectEnvSettingsUC.ImportSettingsToProjectEnv(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
