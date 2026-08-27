package projecthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/projectuc/projectdto"
)

// UpdateProject Updates a project
// @Summary Updates a project
// @Description Updates a project
// @Tags    projects
// @Produce json
// @Id      updateProject
// @Param   projectID path string true "project ID"
// @Param   body body projectdto.UpdateProjectReq true "request data"
// @Success 200 {object} projectdto.UpdateProjectResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID} [put]
func (h *Handler) UpdateProject(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeWrite, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewUpdateProjectReq()
	req.ID = projectID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.UpdateProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
