package projecthandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/projectuc/projectdto"
)

// ListProjectBase Lists projects
// @Summary Lists projects
// @Description Lists projects
// @Tags    projects
// @Produce json
// @Id      listProjectBase
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} projectdto.ListProjectBaseResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/base [get]
func (h *Handler) ListProjectBase(ctx *gin.Context) {
	auth, _, err := h.GetAuth(ctx, base.ActionTypeRead, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewListProjectBaseReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.ListProjectBase(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ListProject Lists projects
// @Summary Lists projects
// @Description Lists projects
// @Tags    projects
// @Produce json
// @Id      listProject
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} projectdto.ListProjectResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects [get]
func (h *Handler) ListProject(ctx *gin.Context) {
	auth, _, err := h.GetAuth(ctx, base.ActionTypeRead, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewListProjectReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.ListProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetProject Gets project details
// @Summary Gets project details
// @Description Gets project details
// @Tags    projects
// @Produce json
// @Id      getProject
// @Param   projectID path string true "project ID"
// @Success 200 {object} projectdto.GetProjectResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID} [get]
func (h *Handler) GetProject(ctx *gin.Context) {
	auth, projectID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := projectdto.NewGetProjectReq()
	req.ID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.projectUC.GetProject(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
