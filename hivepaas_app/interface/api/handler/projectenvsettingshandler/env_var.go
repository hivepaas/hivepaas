package projectenvsettingshandler

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
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvEnvVars
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Success 200 {object} projectsettingsdto.GetProjectEnvVarsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/env-vars [get]
func (h *Handler) GetEnvVars(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewGetProjectEnvVarsReq()
	req.ProjectID = projectID
	req.ProjectEnv = projectEnvID
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
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvEnvVars
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body projectsettingsdto.UpdateProjectEnvVarsReq true "request data"
// @Success 200 {object} projectsettingsdto.UpdateProjectEnvVarsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/env-vars [put]
func (h *Handler) UpdateEnvVars(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewUpdateProjectEnvVarsReq()
	req.ProjectID = projectID
	req.ProjectEnv = projectEnvID
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

// ComputeEnvVars Computes project env vars
// @Summary Computes project env vars
// @Description Computes project env vars
// @Tags    project_env_settings
// @Produce json
// @Id      computeProjectEnvEnvVars
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body projectsettingsdto.ComputeProjectEnvVarsReq true "request data"
// @Success 200 {object} projectsettingsdto.ComputeProjectEnvVarsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/env-vars/compute [post]
func (h *Handler) ComputeEnvVars(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectsettingsdto.NewComputeProjectEnvVarsReq()
	req.ProjectID = projectID
	req.ProjectEnv = projectEnvID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectSettingsUC.ComputeProjectEnvVars(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
