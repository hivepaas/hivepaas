package spechandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// importMaxBodySize caps an import request. A bundle is configuration - an app
// is some hundreds of lines - and arrives base64 inside JSON; this is far above
// any real installation, and the limit the git webhook takes for the same
// reason: the body is read into memory before anything else is known.
const importMaxBodySize = 25 * 1024 * 1024

// ValidateGlobalImport godoc
//
//	@Summary	Plan importing a configuration bundle into the whole installation
//	@Tags		spec
//	@Accept		json
//	@Produce	json
//	@Param		body	body		specdto.ValidateImportReq	true	"bundle, passphrase, selection, options"
//	@Success	200		{object}	specdto.ValidateImportResp
//	@Router		/spec/import/validate [post]
func (h *Handler) ValidateGlobalImport(ctx *gin.Context) {
	h.validateImport(ctx, entity.NewObjectScopeGlobal())
}

// ValidateProjectImport godoc
//
//	@Summary	Plan importing a configuration bundle into one project
//	@Tags		spec
//	@Accept		json
//	@Produce	json
//	@Param		projectID	path		string						true	"Project ID"
//	@Param		body		body		specdto.ValidateImportReq	true	"bundle, passphrase, selection, options"
//	@Success	200			{object}	specdto.ValidateImportResp
//	@Router		/projects/{projectID}/spec/import/validate [post]
func (h *Handler) ValidateProjectImport(ctx *gin.Context) {
	projectID, err := h.ParseStringParam(ctx, "projectID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	h.validateImport(ctx, entity.NewObjectScopeProject(projectID))
}

// ValidateProjectEnvImport godoc
//
//	@Summary	Plan importing a configuration bundle into one project env
//	@Tags		spec
//	@Accept		json
//	@Produce	json
//	@Param		projectID	path		string						true	"Project ID"
//	@Param		projectEnv	path		string						true	"Project env"
//	@Param		body		body		specdto.ValidateImportReq	true	"bundle, passphrase, selection, options"
//	@Success	200			{object}	specdto.ValidateImportResp
//	@Router		/projects/{projectID}/{projectEnv}/spec/import/validate [post]
func (h *Handler) ValidateProjectEnvImport(ctx *gin.Context) {
	projectID, err := h.ParseStringParam(ctx, "projectID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	projectEnv, err := h.ParseStringParam(ctx, "projectEnv")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	h.validateImport(ctx, entity.NewObjectScopeProjectEnv(projectID, projectEnv))
}

func (h *Handler) validateImport(ctx *gin.Context, scope *entity.ObjectScope) {
	accessCheck, err := scopeAccessCheck(scope, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	auth, err := h.authHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, importMaxBodySize)
	req := specdto.NewValidateImportReq()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			err = hperrors.Wrap(hperrors.ErrRequestTooBig).WithParam("MaxSize", importMaxBodySize)
		}
		h.RenderError(ctx, err)
		return
	}
	req.Scope = scope

	resp, err := h.specUC.ValidateImport(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
