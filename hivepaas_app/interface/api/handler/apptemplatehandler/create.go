package apptemplatehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apptemplateuc/apptemplatedto"
)

// CreateAppFromTemplate Creates an app from an app template
// @Summary Creates an app from an app template
// @Description Provisions the app with the template's configuration and queues its first deployment.
// @Tags    app_templates
// @Produce json
// @Id      createAppFromTemplate
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   body body apptemplatedto.CreateAppFromTemplateReq true "request data"
// @Success 201 {object} apptemplatedto.CreateAppFromTemplateResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/from-template [post]
func (h *Handler) CreateAppFromTemplate(ctx *gin.Context) {
	auth, projectID, projectEnvID, _, err := h.GetAuthInEnv(ctx, base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apptemplatedto.NewCreateAppFromTemplateReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appTemplateUC.CreateAppFromTemplate(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}
