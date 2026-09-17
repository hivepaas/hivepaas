package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// GetAppTemplateImageTags Lists the image tags a template version could use
// @Summary Lists the image tags a template version could use
// @Description Reads the registry on demand. The repository comes from the template, never from the request.
// @Tags    app_templates
// @Produce json
// @Id      getAppTemplateImageTags
// @Param   projectID path string true "project ID"
// @Param   templateName path string true "template name"
// @Param   version query string false "template version; empty for the default"
// @Param   variant query string false "template variant; empty for the default"
// @Success 200 {object} apptemplatedto.GetAppTemplateImageTagsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/app-templates/{templateName}/image-tags [get]
func (h *Handler) GetAppTemplateImageTags(ctx *gin.Context) {
	auth, projectID, err := h.GetAuthInProject(ctx, base.ActionTypeRead)
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
	req.ProjectID = projectID
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
