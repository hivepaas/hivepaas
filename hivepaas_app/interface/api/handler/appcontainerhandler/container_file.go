package appcontainerhandler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
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
// @Param   containerId query string false "container ID (optional, auto-picks active container if empty)"
// @Param   path query string true "file/dir path in container"
// @Param   isDir query boolean false "is directory"
// @Param   compressionFormat query string false "compression format (gzip, zstd)"
// @Success 200 {file} binary
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/container/file-download [get]
func (h *Handler) DownloadFileFromContainer(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeExecute)
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

	ctx.Header("Content-Disposition", fmt.Sprintf(
		`attachment; filename="%s"; filename*=UTF-8''%s`,
		resp.FileName,
		url.QueryEscape(resp.FileName),
	))
	ctx.Header("Content-Type", resp.ContentType)
	if resp.FileSize > 0 {
		ctx.Header("Content-Length", strconv.FormatInt(resp.FileSize, 10))
	}

	_, _ = io.Copy(ctx.Writer, resp.Reader)
}

// UploadFileToContainer Uploads a file or archive into container
// @Summary Uploads a file or archive into container
// @Description Uploads a file or archive into container, with optional archive extraction
// @Tags    apps
// @Accept  multipart/form-data
// @Produce json
// @Id      uploadFileToAppContainer
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   nodeId formData string false "node ID"
// @Param   containerId formData string false "container ID (optional, auto-picks active container if empty)"
// @Param   path formData string true "file/dir path in container"
// @Param   extract formData boolean false "extract archive into path"
// @Param   compressionFormat formData string false "compression format (gzip, zstd, zip, tar)"
// @Param   overwrite formData boolean false "allow overwrite (default: true)"
// @Param   file formData file true "file to upload"
// @Success 200 {object} appcontainerdto.UploadFileToContainerResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/container/file-upload [post]
func (h *Handler) UploadFileToContainer(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appcontainerdto.NewUploadFileToContainerReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID

	if _, err = ctx.MultipartForm(); err != nil {
		h.RenderError(ctx, err)
		return
	}
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		h.RenderError(ctx, apperrors.NewArgumentInvalid("file").WithMsgLog("file is required: %v", err))
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		h.RenderError(ctx, apperrors.Wrap(err))
		return
	}
	defer file.Close()

	req.FileName = fileHeader.Filename
	req.FileSize = fileHeader.Size
	req.FileContent = file

	resp, err := h.appContainerUC.UploadFileToContainer(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
