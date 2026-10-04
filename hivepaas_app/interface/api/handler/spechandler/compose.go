package spechandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// ValidateCompose godoc
//
//	@Summary	Plan creating a project from a Docker Compose file
//	@Tags		Configuration specs
//	@Accept		json
//	@Produce	json
//	@Param		body	body		specdto.ValidateComposeReq	true	"the compose file, its .env and files, the choices"
//	@Success	200		{object}	specdto.ValidateComposeResp
//	@Router		/projects/from-compose/validate [post]
func (h *Handler) ValidateCompose(ctx *gin.Context) {
	req := specdto.NewValidateComposeReq()
	auth, ok := h.readComposeReq(ctx, req)
	if !ok {
		return
	}
	resp, err := h.specUC.ValidateCompose(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// ApplyCompose godoc
//
//	@Summary	Create a project from a Docker Compose file
//	@Tags		Configuration specs
//	@Accept		json
//	@Produce	json
//	@Param		body	body		specdto.ApplyComposeReq	true	"validate's body, planHash, acceptIssues"
//	@Success	200		{object}	specdto.ApplyComposeResp
//	@Router		/projects/from-compose/apply [post]
func (h *Handler) ApplyCompose(ctx *gin.Context) {
	req := specdto.NewApplyComposeReq()
	auth, ok := h.readComposeReq(ctx, req)
	if !ok {
		return
	}
	resp, err := h.specUC.ApplyCompose(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// readComposeReq checks the caller may create a project - the gate POST
// /projects has - and reads a body of at most importMaxBodySize into req. It
// renders the error and answers false when either fails.
func (h *Handler) readComposeReq(ctx *gin.Context, req any) (*basedto.Auth, bool) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
	})
	if err != nil {
		h.RenderError(ctx, err)
		return nil, false
	}
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, importMaxBodySize)
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			err = hperrors.Wrap(hperrors.ErrRequestTooBig).WithParam("MaxSize", importMaxBodySize)
		}
		h.RenderError(ctx, err)
		return nil, false
	}
	return auth, true
}
