package appsettingshandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

// ListDataFile Lists data files of an app
// @Summary Lists data files of an app
// @Description Lists data files of an app
// @Tags    app_settings
// @Produce json
// @Id      listAppDataFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} filedto.ListFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/data-files [get]
func (h *Handler) ListDataFile(ctx *gin.Context) {
	auth, _, appID, err := h.GetAuth(ctx, base.ActionTypeRead, true)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewListFileReq()
	req.ObjectID = appID
	if err := h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.ListFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetDataFile Gets a data file of an app
// @Summary Gets a data file of an app
// @Description Gets a data file of an app
// @Tags    app_settings
// @Produce json
// @Id      getAppDataFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "file ID"
// @Success 200 {object} filedto.GetFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/data-files/{itemID} [get]
func (h *Handler) GetDataFile(ctx *gin.Context) {
	auth, _, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewGetFileReq()
	req.ID = itemID
	req.ObjectID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.GetFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetDataFileDownloadURL Gets download url of a data file
// @Summary Gets download url of a data file
// @Description Gets download url of a data file
// @Tags    app_settings
// @Produce json
// @Id      getAppDataFileDownloadURL
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "file ID"
// @Success 200 {object} filedto.GetFileDownloadURLResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/data-files/{itemID}/download-url [get]
func (h *Handler) GetDataFileDownloadURL(ctx *gin.Context) {
	auth, _, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewGetFileDownloadURLReq()
	req.ID = itemID
	req.ObjectID = appID
	req.Expiration = timeutil.Duration(time.Minute)
	req.CloudPresign = true
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.GetFileDownloadURL(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteDataFile Deletes a data file of an app
// @Summary Deletes a data file of an app
// @Description Deletes a data file of an app
// @Tags    app_settings
// @Produce json
// @Id      deleteAppDataFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "file ID"
// @Success 200 {object} filedto.DeleteFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/data-files/{itemID} [delete]
func (h *Handler) DeleteDataFile(ctx *gin.Context) {
	auth, _, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewDeleteFileReq()
	req.ID = itemID
	req.ObjectID = appID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.DeleteFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
