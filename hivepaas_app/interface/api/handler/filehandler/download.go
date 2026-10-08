package filehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

// DownloadFile Downloads a file
// @Summary Downloads a file
// @Description Downloads a file
// @Tags    Files
// @Produce application/octet-stream
// @Id      downloadFile
// @Param   fileID path string true "file ID"
// @Param   token query string false "`token=<download token>`, from a download URL, in place of a session"
// @Param   viewInline query bool false "`viewInline=true` to show the file in the browser rather than save it"
// @Success 200
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
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
