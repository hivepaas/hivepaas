package spechandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// reportHeader carries a summary of what the export skipped or could not
// resolve: counts, not detail.
//
// The detail is inside the bundle as report.yaml. It cannot travel here -
// measured against a development installation of three projects and five apps
// the full report already reached 6.8 KB, inside nginx's 8 KB default for the
// entire header block. A real installation would exceed it and the failure
// would be a truncated header or a proxy error rather than anything legible.
//
// The header must also be listed in the CORS ExposeHeaders, or a dashboard
// running on its own dev origin cannot read it - see middleware/cors.
const reportHeader = "X-HivePaaS-Spec-Report"

// ExportGlobalSpec godoc
//
//	@Summary	Export the whole installation as a configuration spec
//	@Tags		spec
//	@Produce	application/gzip
//	@Param		body	body	specdto.ExportSpecReq	false	"secretsMode and passphrase"
//	@Success	200
//	@Router		/spec/export [post]
func (h *Handler) ExportGlobalSpec(ctx *gin.Context) {
	req := specdto.NewExportSpecReq()
	req.Scope = entity.NewObjectScopeGlobal()
	h.export(ctx, req)
}

// ExportProjectSpec godoc
//
//	@Summary	Export one project as a configuration spec
//	@Tags		spec
//	@Produce	application/gzip
//	@Param		projectID	path	string	true	"Project ID"
//	@Success	200
//	@Router		/projects/{projectID}/spec/export [post]
func (h *Handler) ExportProjectSpec(ctx *gin.Context) {
	req := specdto.NewExportSpecReq()
	projectID, err := h.ParseStringParam(ctx, "projectID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req.Scope = entity.NewObjectScopeProject(projectID)
	h.export(ctx, req)
}

// ExportProjectEnvSpec godoc
//
//	@Summary	Export one project env as a configuration spec
//	@Tags		spec
//	@Produce	application/gzip
//	@Param		projectID	path	string	true	"Project ID"
//	@Param		projectEnv	path	string	true	"Project env"
//	@Success	200
//	@Router		/projects/{projectID}/{projectEnv}/spec/export [post]
func (h *Handler) ExportProjectEnvSpec(ctx *gin.Context) {
	req := specdto.NewExportSpecReq()
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
	req.Scope = entity.NewObjectScopeProjectEnv(projectID, projectEnv)
	h.export(ctx, req)
}

// ExportAppSpec godoc
//
//	@Summary	Export one app as a configuration spec
//	@Tags		spec
//	@Produce	application/gzip
//	@Param		projectID	path	string	true	"Project ID"
//	@Param		projectEnv	path	string	true	"Project env"
//	@Param		appID		path	string	true	"App ID"
//	@Success	200
//	@Router		/projects/{projectID}/{projectEnv}/apps/{appID}/spec/export [post]
func (h *Handler) ExportAppSpec(ctx *gin.Context) {
	req := specdto.NewExportSpecReq()
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
	appID, err := h.ParseStringParam(ctx, "appID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	req.Scope = entity.NewObjectScopeApp(appID, "", projectID, projectEnv)
	h.export(ctx, req)
}

// accessCheckForScope picks the check that matches what is being exported.
//
// The same check for every route would be wrong in both directions at once:
// gating a single project on the system module locks out a project member who
// can already read every setting the export would contain, and lets anyone with
// system read export a project they have no access to. Each scope is gated the
// way every other handler gates it - ProjectAccessCheck for a project or env,
// AppAccessCheck for an app.
//
// The global export keeps the system-module check because it genuinely crosses
// every project, which is the same reason the audit log listing uses it.
func accessCheckForScope(req *specdto.ExportSpecReq) (permission.AccessCheck, error) {
	read := permission.BaseAccessCheck{Action: base.ActionTypeRead}

	switch req.Scope.ScopeType {
	case base.ObjectScopeProject:
		return &permission.ProjectAccessCheck{
			BaseAccessCheck: read,
			ProjectID:       req.Scope.ProjectID,
		}, nil

	case base.ObjectScopeProjectEnv:
		return &permission.ProjectAccessCheck{
			BaseAccessCheck: read,
			ProjectID:       req.Scope.ProjectID,
			ProjectEnv:      &req.Scope.ProjectEnvID,
		}, nil

	case base.ObjectScopeApp:
		return &permission.AppAccessCheck{
			BaseAccessCheck: read,
			ProjectID:       req.Scope.ProjectID,
			ProjectEnv:      req.Scope.ProjectEnvID,
			AppID:           req.Scope.AppID,
		}, nil

	case base.ObjectScopeGlobal:
		return &permission.GeneralResourceAccessCheck{
			BaseAccessCheck: read,
			Module:          base.ResourceModuleSystem,
			ResourceType:    base.ResourceTypeSetting,
		}, nil

	case base.ObjectScopeUser, base.ObjectScopeHivepaas:
	}
	return nil, hperrors.Wrap(hperrors.ErrObjectScopeInvalid)
}

func (h *Handler) export(ctx *gin.Context, req *specdto.ExportSpecReq) {
	accessCheck, err := accessCheckForScope(req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	auth, err := h.authHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// The body, not the query string: the passphrase must not reach an access
	// log. See specdto.ExportSpecReq.
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.specUC.ExportSpec(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer resp.Data.Content.Close()

	if resp.Summary != nil {
		if encoded, encodeErr := json.Marshal(resp.Summary); encodeErr == nil {
			ctx.Header(reportHeader, string(encoded))
		}
	}

	ctx.DataFromReader(http.StatusOK, resp.Data.ContentLength, resp.Data.ContentType,
		resp.Data.Content, resp.Data.ExtraHeaders)
}
