package spechandler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
	req := specdto.NewValidateImportReq()
	auth, ok := h.readImportReq(ctx, scope, req)
	if !ok {
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

// ApplyGlobalImport godoc
//
//	@Summary	Import a configuration bundle into the whole installation
//	@Tags		spec
//	@Accept		json
//	@Produce	json
//	@Param		body	body		specdto.ApplyImportReq	true	"validate's body, planHash, acceptIssues"
//	@Success	200		{object}	specdto.ApplyImportResp
//	@Router		/spec/import/apply [post]
func (h *Handler) ApplyGlobalImport(ctx *gin.Context) {
	h.applyImport(ctx, entity.NewObjectScopeGlobal())
}

// ApplyProjectImport godoc
//
//	@Summary	Import a configuration bundle into one project
//	@Tags		spec
//	@Accept		json
//	@Produce	json
//	@Param		projectID	path		string					true	"Project ID"
//	@Param		body		body		specdto.ApplyImportReq	true	"validate's body, planHash, acceptIssues"
//	@Success	200			{object}	specdto.ApplyImportResp
//	@Router		/projects/{projectID}/spec/import/apply [post]
func (h *Handler) ApplyProjectImport(ctx *gin.Context) {
	projectID, err := h.ParseStringParam(ctx, "projectID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	h.applyImport(ctx, entity.NewObjectScopeProject(projectID))
}

// ApplyProjectEnvImport godoc
//
//	@Summary	Import a configuration bundle into one project env
//	@Tags		spec
//	@Accept		json
//	@Produce	json
//	@Param		projectID	path		string					true	"Project ID"
//	@Param		projectEnv	path		string					true	"Project env"
//	@Param		body		body		specdto.ApplyImportReq	true	"validate's body, planHash, acceptIssues"
//	@Success	200			{object}	specdto.ApplyImportResp
//	@Router		/projects/{projectID}/{projectEnv}/spec/import/apply [post]
func (h *Handler) ApplyProjectEnvImport(ctx *gin.Context) {
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
	h.applyImport(ctx, entity.NewObjectScopeProjectEnv(projectID, projectEnv))
}

func (h *Handler) applyImport(ctx *gin.Context, scope *entity.ObjectScope) {
	req := specdto.NewApplyImportReq()
	auth, ok := h.readImportReq(ctx, scope, req)
	if !ok {
		return
	}
	req.Scope = scope

	resp, err := h.specUC.ApplyImport(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// readImportReq checks the caller may write at the route's scope, and reads an
// import body of at most importMaxBodySize into req. It renders the error and
// answers false when either fails.
func (h *Handler) readImportReq(ctx *gin.Context, scope *entity.ObjectScope, req any) (*basedto.Auth, bool) {
	accessCheck, err := scopeAccessCheck(scope, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return nil, false
	}
	auth, err := h.authHandler.GetCurrentAuth(ctx, accessCheck)
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
