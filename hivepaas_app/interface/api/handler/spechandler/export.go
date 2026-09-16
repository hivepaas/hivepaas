package spechandler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

// reportHeader carries what the export skipped or could not resolve. The body
// is the archive itself, so there is nowhere else to put it.
const reportHeader = "X-HivePaaS-Spec-Report"

// ExportGlobalSpec godoc
//
//	@Summary	Export the whole installation as a configuration spec
//	@Tags		spec
//	@Produce	application/gzip
//	@Param		secretsMode	query	string	false	"omit (default), encrypted or plaintext"
//	@Param		passphrase	query	string	false	"required when secretsMode is encrypted"
//	@Success	200
//	@Router		/spec/export [get]
func (h *Handler) ExportGlobalSpec(ctx *gin.Context) {
	h.export(ctx, specdto.NewExportSpecReq())
}

// ExportProjectSpec godoc
//
//	@Summary	Export one project as a configuration spec
//	@Tags		spec
//	@Produce	application/gzip
//	@Param		projectID	path	string	true	"Project ID"
//	@Success	200
//	@Router		/projects/{projectID}/spec/export [get]
func (h *Handler) ExportProjectSpec(ctx *gin.Context) {
	req := specdto.NewExportSpecReq()
	req.ProjectID = ctx.Param("projectID")
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
//	@Router		/projects/{projectID}/{projectEnv}/spec/export [get]
func (h *Handler) ExportProjectEnvSpec(ctx *gin.Context) {
	req := specdto.NewExportSpecReq()
	req.ProjectID = ctx.Param("projectID")
	req.ProjectEnvID = projecthelper.CalcProjectEnvID(req.ProjectID, ctx.Param("projectEnv"))
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
//	@Router		/projects/{projectID}/{projectEnv}/apps/{appID}/spec/export [get]
func (h *Handler) ExportAppSpec(ctx *gin.Context) {
	req := specdto.NewExportSpecReq()
	req.ProjectID = ctx.Param("projectID")
	req.ProjectEnvID = projecthelper.CalcProjectEnvID(req.ProjectID, ctx.Param("projectEnv"))
	req.AppID = ctx.Param("appID")
	h.export(ctx, req)
}

func (h *Handler) export(ctx *gin.Context, req *specdto.ExportSpecReq) {
	auth, err := h.authHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeSetting,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.specUC.ExportSpec(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer resp.Data.Content.Close()

	if resp.Report != nil && len(resp.Report.Issues) > 0 {
		if encoded, encodeErr := json.Marshal(resp.Report); encodeErr == nil {
			ctx.Header(reportHeader, string(encoded))
		}
	}

	ctx.DataFromReader(http.StatusOK, resp.Data.ContentLength, resp.Data.ContentType,
		resp.Data.Content, resp.Data.ExtraHeaders)
}
