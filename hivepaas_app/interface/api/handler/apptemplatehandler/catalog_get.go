package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateCatalog Gets the app template catalog
// @Summary Gets the app template catalog
// @Description The source, its revision, and the categories and tags the template list can be filtered by.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplateCatalog
// @Param   projectID path string true "project ID"
// @Success 200 {object} apptemplatedto.GetAppTemplateCatalogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/app-template-catalog [get]
func (h *Handler) GetAppTemplateCatalog(ctx *gin.Context) {
	auth, projectID, err := h.GetAuthInProject(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateCatalogReq()
	req.ProjectID = projectID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplateCatalog(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
