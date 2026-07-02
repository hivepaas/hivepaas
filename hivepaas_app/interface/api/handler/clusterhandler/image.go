package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/imageuc/imagedto"
)

// ListImage Lists cluster images
// @Summary Lists cluster images
// @Description Lists cluster images
// @Tags    cluster_images
// @Produce json
// @Id      listClusterImage
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} imagedto.ListImageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/images [get]
func (h *Handler) ListImage(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeImage, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagedto.NewListImageReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.imageUC.ListImage(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetImage Gets image details
// @Summary Gets image details
// @Description Gets image details
// @Tags    cluster_images
// @Produce json
// @Id      getClusterImage
// @Param   imageID path string true "image ID"
// @Success 200 {object} imagedto.GetImageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/images/{imageID} [get]
func (h *Handler) GetImage(ctx *gin.Context) {
	auth, imageID, err := h.getAuth(ctx, base.ResourceTypeImage, base.ActionTypeRead, "imageID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagedto.NewGetImageReq()
	req.ImageID = imageID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.imageUC.GetImage(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetImageInspection Gets image details
// @Summary Gets image details
// @Description Gets image details
// @Tags    cluster_images
// @Produce json
// @Id      getClusterImageInspection
// @Param   imageID path string true "image ID"
// @Success 200 {object} imagedto.GetImageInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/images/{imageID}/inspect [get]
func (h *Handler) GetImageInspection(ctx *gin.Context) {
	auth, imageID, err := h.getAuth(ctx, base.ResourceTypeImage, base.ActionTypeRead, "imageID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagedto.NewGetImageInspectionReq()
	req.ImageID = imageID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.imageUC.GetImageInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateImage Creates an image
// @Summary Creates an image
// @Description Creates an image
// @Tags    cluster_images
// @Produce json
// @Id      createClusterImage
// @Param   body body imagedto.CreateImageReq true "request data"
// @Success 200 {object} imagedto.CreateImageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/images [post]
func (h *Handler) CreateImage(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeImage, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagedto.NewCreateImageReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.imageUC.CreateImage(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteImage Deletes an image
// @Summary Deletes an image
// @Description Deletes an image
// @Tags    cluster_images
// @Produce json
// @Id      deleteClusterImage
// @Param   imageID path string true "image ID"
// @Success 200 {object} imagedto.DeleteImageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/images/{imageID} [delete]
func (h *Handler) DeleteImage(ctx *gin.Context) {
	auth, imageID, err := h.getAuth(ctx, base.ResourceTypeImage, base.ActionTypeDelete, "imageID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagedto.NewDeleteImageReq()
	req.ImageID = imageID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.imageUC.DeleteImage(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
