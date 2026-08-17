package appcontainerhandler

import (
	"io"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
)

// DownloadFileFromContainer Downloads a file or directory from container
// @Summary Downloads a file or directory from container
// @Description Downloads a file or directory from container
// @Tags    apps
// @Produce octet-stream
// @Id      downloadFileFromAppContainer
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   nodeId query string false "node ID"
// @Param   containerId query string true "container ID"
// @Param   path query string true "file/dir path in container"
// @Param   isDir query boolean false "is directory"
// @Param   compressionFormat query string false "compression format (gzip, zstd)"
// @Success 200 {file} binary
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/container/file-download [get]
func (h *Handler) DownloadFileFromContainer(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appcontainerdto.NewDownloadFileFromContainerReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appContainerUC.DownloadFileFromContainer(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	defer resp.Reader.Close()

	ctx.Header("Content-Disposition", "attachment; filename="+resp.FileName)
	ctx.Header("Content-Type", resp.ContentType)
	if resp.FileSize > 0 {
		ctx.Header("Content-Length", strconv.FormatInt(resp.FileSize, 10))
	}

	_, _ = io.Copy(ctx.Writer, resp.Reader)
}
