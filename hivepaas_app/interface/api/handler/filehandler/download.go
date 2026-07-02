package filehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

// DownloadFile Downloads a file
// @Summary Downloads a file
// @Description Downloads a file
// @Tags    files
// @Produce application/octet-stream
// @Id      downloadFile
// @Param   fileID path string true "file ID"
// @Success 200
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /files/{fileID}/download [get]
func (h *Handler) DownloadFile(ctx *gin.Context) {
	fileID, err := h.ParseStringParam(ctx, "fileID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// NOTE: `auth` will be handled within the use case along with other params
	auth, _ := h.authHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)

	req := filedto.NewDownloadFileReq()
	req.ID = fileID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.DownloadFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	data := resp.Data
	defer data.Content.Close()

	ctx.DataFromReader(http.StatusOK, data.ContentLength, data.ContentType, data.Content, data.ExtraHeaders)
}
