package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateIcon Gets an app template icon
// @Summary Gets an app template icon
// @Description Public: an <img> cannot send the Authorization header. Only icons the current index lists are served.
// @Tags    app_templates
// @Produce image/svg+xml,image/png
// @Id      getAppTemplateIcon
// @Param   sha256 path string true "icon sha256"
// @Success 200
// @Failure 404 {object} hperrors.ErrorInfo
// @Router  /app-templates/icons/{sha256} [get]
func (h *Handler) GetAppTemplateIcon(ctx *gin.Context) {
	sha256Hex, err := h.ParseStringParam(ctx, "sha256")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateIconReq()
	req.SHA256 = sha256Hex
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplateIcon(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// An SVG can carry script. sandbox keeps it inert even when opened directly,
	// and nosniff keeps the declared type the only one a browser uses.
	ctx.Header("Content-Security-Policy", "sandbox")
	ctx.Header("X-Content-Type-Options", "nosniff")
	ctx.Header("Cache-Control", "public, max-age=31536000, immutable")
	ctx.Data(http.StatusOK, resp.ContentType, resp.Content)
}
