package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// ListAppTemplates Lists app templates
// @Summary Lists app templates
// @Description Lists the templates a project can create apps from, a page at a time, ordered by name.
// @Description The categories and tags to filter by come from the app template catalog.
// @Tags    app_templates
// @Produce json
// @Id      listAppTemplates
// @Param   projectID path string true "project ID"
// @Param   category query string false "categories, comma separated; a parent matches every child"
// @Param   tag query string false "tags, comma separated; a template carrying any of them matches"
// @Param   search query string false "matches name, title, tagline, tags and aliases, ignoring case"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Success 200 {object} apptemplatedto.ListAppTemplatesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/app-templates [get]
func (h *Handler) ListAppTemplates(ctx *gin.Context) {
	auth, projectID, err := h.GetAuthInProject(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewListAppTemplatesReq()
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.ListAppTemplates(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
