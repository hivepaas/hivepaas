package apppreviewhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/apppreviewuc/apppreviewdto"
)

// CreateAppPreview Creates preview for an app
// @Summary Creates preview for an app
// @Description Creates preview for an app
// @Tags    app_previews
// @Produce json
// @Id      createAppPreview
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body apppreviewdto.CreatePreviewReq true "request data"
// @Success 201 {object} apppreviewdto.CreatePreviewResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/previews [post]
func (h *Handler) CreateAppPreview(ctx *gin.Context) {
	auth, projectID, appID, _, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apppreviewdto.NewCreatePreviewReq()
	req.ProjectID = projectID
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
// @Param   appID path string true "app ID"
// @Param   body body apppreviewdto.PrepareCreatePreviewReq true "request data"
// @Success 200 {object} apppreviewdto.PrepareCreatePreviewResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/previews/prepare [post]
func (h *Handler) PrepareCreateAppPreview(ctx *gin.Context) {
	auth, projectID, appID, _, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := apppreviewdto.NewPrepareCreatePreviewReq()
	req.ProjectID = projectID
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
