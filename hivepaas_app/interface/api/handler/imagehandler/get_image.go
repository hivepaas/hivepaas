package imagehandler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/binobjectuc/binobjectdto"
)

// GetPublicImage Gets an image file
// @Summary Gets an image file
// @Description Gets an image file
// @Tags    images
// @Produce json
// @Id      getPublicImage
// @Param   imageID path string true "image ID"
// @Success 200
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /images/{imageID} [get]
func (h *Handler) GetPublicImage(ctx *gin.Context) {
	imageID, err := h.ParseStringParam(ctx, "imageID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := binobjectdto.NewGetBinObjectDataReq()
	req.ID, _, _ = strings.Cut(imageID, "-")
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.binObjectUC.GetBinObjectData(h.RequestCtx(ctx), nil, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	data := resp.Data
	defer data.Content.Close()

	ctx.DataFromReader(http.StatusOK, data.ContentLength, data.ContentType, data.Content, data.ExtraHeaders)
}
