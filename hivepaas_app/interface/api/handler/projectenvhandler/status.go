package projectenvhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectenvuc/projectenvdto"
)

// UpdateProjectEnvStatus Updates project env status
// @Summary Updates project env status
// @Description Updates project env status
// @Tags    project_envs
// @Produce json
// @Id      updateProjectEnvStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body projectenvdto.UpdateProjectEnvStatusReq true "request data"
// @Success 200 {object} projectenvdto.UpdateProjectEnvStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/status [put]
func (h *Handler) UpdateProjectEnvStatus(ctx *gin.Context) {
	auth, projectID, projectEnvID, err := h.GetEnvAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectenvdto.NewUpdateProjectEnvStatusReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectEnvUC.UpdateProjectEnvStatus(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
