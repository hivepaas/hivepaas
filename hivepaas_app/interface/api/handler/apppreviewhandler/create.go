package apppreviewhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc/apppreviewdto"
)

// CreateAppPreview Creates preview for an app
// @Summary Creates preview for an app
// @Description Creates preview for an app
// @Tags    app_previews
// @Produce json
// @Id      createAppPreview
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body apppreviewdto.CreatePreviewReq true "request data"
// @Success 201 {object} apppreviewdto.CreatePreviewResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/previews [post]
func (h *Handler) CreateAppPreview(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, _, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apppreviewdto.NewCreatePreviewReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appPreviewUC.CreatePreview(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// PrepareCreateAppPreview Prepares to create a preview for an app
// @Summary Prepares to create a preview for an app
// @Description Prepares to create a preview for an app
// @Tags    app_previews
// @Produce json
// @Id      prepareCreateAppPreview
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body apppreviewdto.PrepareCreatePreviewReq true "request data"
// @Success 200 {object} apppreviewdto.PrepareCreatePreviewResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/previews/prepare [post]
func (h *Handler) PrepareCreateAppPreview(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, _, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apppreviewdto.NewPrepareCreatePreviewReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appPreviewUC.PrepareCreatePreview(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
