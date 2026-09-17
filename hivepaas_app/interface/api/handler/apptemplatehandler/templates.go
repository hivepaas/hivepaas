package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// ListAppTemplates Lists app templates
// @Summary Lists app templates
// @Description Lists the templates a project can create apps from, a page at a time, ordered by name.
// @Description The categories and tags to filter by come from the app template catalog.
// @Tags    app_templates
// @Produce json
// @Id      listAppTemplates
// @Param   category query string false "categories, comma separated; a parent matches every child"
// @Param   tag query string false "tags, comma separated; a template carrying any of them matches"
// @Param   search query string false "matches name, title, tagline, tags and aliases, ignoring case"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Success 200 {object} apptemplatedto.ListAppTemplatesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /app-templates [get]
func (h *Handler) ListAppTemplates(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewListAppTemplatesReq()
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

// GetAppTemplate Gets an app template
// @Summary Gets an app template
// @Description Gets one app template: its description, versions, variants and parameters.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplate
// @Param   templateName path string true "template name"
// @Success 200 {object} apptemplatedto.GetAppTemplateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /app-templates/{templateName} [get]
func (h *Handler) GetAppTemplate(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	name, err := h.ParseStringParam(ctx, "templateName")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateReq()
	req.Name = name
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplate(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppTemplateImageTags Lists the image tags a template version could use
// @Summary Lists the image tags a template version could use
// @Description Reads the registry on demand. The repository comes from the template, never from the request.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplateImageTags
// @Param   templateName path string true "template name"
// @Param   version query string false "template version; empty for the default"
// @Param   variant query string false "template variant; empty for the default"
// @Success 200 {object} apptemplatedto.GetAppTemplateImageTagsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Router  /app-templates/{templateName}/image-tags [get]
func (h *Handler) GetAppTemplateImageTags(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	name, err := h.ParseStringParam(ctx, "templateName")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateImageTagsReq()
	req.Name = name
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.GetAppTemplateImageTags(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetAppTemplateIcon Gets an app template icon
// @Summary Gets an app template icon
// @Description Public: an <img> cannot send the Authorization header. Only icons the current index lists are served.
// @Tags    app_templates
// @Produce image/svg+xml,image/png
// @Id      getAppTemplateIcon
// @Param   file path string true "icon file name, such as postgres.3a18fec853.svg"
// @Success 200
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Router  /app-templates/icons/{file} [get]
func (h *Handler) GetAppTemplateIcon(ctx *gin.Context) {
	file, err := h.ParseStringParam(ctx, "file")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewGetAppTemplateIconReq()
	req.File = file
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
