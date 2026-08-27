package apphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// CreateApp Creates a new app
// @Summary Creates a new app
// @Description Creates a new app
// @Tags    apps
// @Produce json
// @Id      createApp
// @Param   projectID path string true "project ID"
// @Param   body body appdto.CreateAppReq true "request data"
// @Success 201 {object} appdto.CreateAppResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/apps [post]
func (h *Handler) CreateApp(ctx *gin.Context) {
	auth, projectID, projectEnvID, _, err := h.GetAuthInEnv(ctx, base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewCreateAppReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.CreateApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}
