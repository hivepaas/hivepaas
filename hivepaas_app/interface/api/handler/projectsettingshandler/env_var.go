package projectsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectsettingsuc/projectsettingsdto"
)

// GetEnvVars Gets project env vars
// @Summary Gets project env vars
// @Description Gets project env vars
// @Tags    project_settings
// @Produce json
// @Id      getProjectEnvVars
// @Param   projectID path string true "project ID"
// @Success 200 {object} projectsettingsdto.GetProjectEnvVarsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/env-vars [get]
func (h *Handler) GetEnvVars(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewGetProjectEnvVarsReq()
	req.ProjectID = projectID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.GetProjectEnvVars(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateEnvVars Updates project env vars
// @Summary Updates project env vars
// @Description Updates project env vars
// @Tags    project_settings
// @Produce json
// @Id      updateProjectEnvVars
// @Param   projectID path string true "project ID"
// @Param   body body projectsettingsdto.UpdateProjectEnvVarsReq true "request data"
// @Success 200 {object} projectsettingsdto.UpdateProjectEnvVarsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/env-vars [put]
func (h *Handler) UpdateEnvVars(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewUpdateProjectEnvVarsReq()
	req.ProjectID = projectID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.UpdateProjectEnvVars(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
