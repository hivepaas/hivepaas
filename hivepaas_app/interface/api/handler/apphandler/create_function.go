package apphandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
)

// CreateFunction Creates a function
// @Summary Creates a function
// @Description Creates an app of kind function from its source, and queues its first deployment.
// @Tags    Apps
// @Produce json
// @Id      createFunction
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   body body appdto.CreateFunctionReq true "request data"
// @Success 201 {object} appdto.CreateFunctionResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/function [post]
func (h *Handler) CreateFunction(ctx *gin.Context) {
	auth, projectID, projectEnvID, _, err := h.GetAuthInEnv(ctx, base.ActionTypeWrite, false)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appdto.NewCreateFunctionReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appUC.CreateFunction(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}
