package projecthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

// UpdateProjectStatus Updates project status
// @Summary Updates project status
// @Description Updates project status
// @Tags    projects
// @Produce json
// @Id      updateProjectStatus
// @Param   projectID path string true "project ID"
// @Param   body body projectdto.UpdateProjectStatusReq true "request data"
// @Success 200 {object} projectdto.UpdateProjectStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/status [put]
func (h *Handler) UpdateProjectStatus(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewUpdateProjectStatusReq()
	req.ID = projectID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.UpdateProjectStatus(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
