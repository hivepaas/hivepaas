package projectenvhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvuc/projectenvdto"
)

// DeleteProjectEnv Deletes a project env
// @Summary Deletes a project env
// @Description Deletes a project env
// @Tags    project_envs
// @Produce json
// @Id      deleteProjectEnv
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Success 200 {object} projectenvdto.DeleteProjectEnvResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv} [delete]
func (h *Handler) DeleteProjectEnv(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetEnvAuth(ctx, base.ActionTypeDelete, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectenvdto.NewDeleteProjectEnvReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectEnvUC.DeleteProjectEnv(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
