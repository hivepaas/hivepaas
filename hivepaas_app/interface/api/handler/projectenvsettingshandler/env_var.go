package projectenvsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvsettingsuc/projectenvsettingsdto"
)

// GetEnvVars Gets project env vars
// @Summary Gets project env vars
// @Description Gets project env vars
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvEnvVars
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Success 200 {object} projectenvsettingsdto.GetProjectEnvEnvVarsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/env-vars [get]
func (h *Handler) GetEnvVars(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectenvsettingsdto.NewGetProjectEnvEnvVarsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectEnvSettingsUC.GetProjectEnvEnvVars(h.RequestCtx(ctx), auth, req)
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
// @Param   body body projectenvsettingsdto.UpdateProjectEnvEnvVarsReq true "request data"
// @Success 200 {object} projectenvsettingsdto.UpdateProjectEnvEnvVarsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/env-vars [put]
func (h *Handler) UpdateEnvVars(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectenvsettingsdto.NewUpdateProjectEnvEnvVarsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectEnvSettingsUC.UpdateProjectEnvEnvVars(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// BuildEnvVars Computes project env vars
// @Summary Computes project env vars
// @Description Computes project env vars
// @Tags    project_env_settings
// @Produce json
// @Id      computeProjectEnvEnvVars
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body projectenvsettingsdto.ComputeProjectEnvEnvVarsReq true "request data"
// @Success 200 {object} projectenvsettingsdto.ComputeProjectEnvEnvVarsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/env-vars/compute [post]
func (h *Handler) BuildEnvVars(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectenvsettingsdto.NewComputeProjectEnvEnvVarsReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectEnvSettingsUC.ComputeProjectEnvEnvVars(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
