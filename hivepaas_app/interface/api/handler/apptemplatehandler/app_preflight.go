package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// PreflightAppFromTemplate Reports what creating an app from a template would run into
// @Summary Reports what creating an app from a template would run into
// @Description Renders the request and reports which of its apps would be given a directory that already holds data.
// @Tags    app_templates
// @Produce json
// @Id      preflightAppFromTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   body body apptemplatedto.PreflightAppFromTemplateReq true "request data"
// @Success 200 {object} apptemplatedto.PreflightAppFromTemplateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/from-template/preflight [post]
func (h *Handler) PreflightAppFromTemplate(ctx *gin.Context) {
	// The same permission the creation needs: this answers with what is on the
	// project's volumes, which is not something a reader may ask about.
	auth, projectID, projectEnvID, _, err := h.GetAuthInEnv(ctx, base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewPreflightAppFromTemplateReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.PreflightAppFromTemplate(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
