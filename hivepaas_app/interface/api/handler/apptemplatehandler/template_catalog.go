package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateCatalog Gets the app template catalog
// @Summary Gets the app template catalog
// @Description The source, its revision, and the categories and tags the template list can be filtered by.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplateCatalog
// @Success 200 {object} apptemplatedto.GetAppTemplateCatalogResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /app-templates/catalog [get]
func (h *Handler) GetAppTemplateCatalog(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateCatalogReq()
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
