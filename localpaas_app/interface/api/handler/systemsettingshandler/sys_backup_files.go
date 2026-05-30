package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/fileuc/filedto"
)

const (
	defaultUsePresignURLOnFileSize = 10 * 1024 * 1024 // 10MB
)

// ListBackupFiles Lists backup files
// @Summary Lists backup files
// @Description Lists backup files
// @Tags    system_settings
// @Produce json
// @Id      listSystemBackupFiles
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} filedto.ListFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/backup/files [get]
func (h *Handler) ListBackupFiles(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemBackup,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewListFileReq()
	req.Types = []base.FileType{base.FileTypeSystemBackup}
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.FileUC.ListFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetBackupFile Gets backup file
// @Summary Gets backup file
// @Description Gets backup file
// @Tags    system_settings
// @Produce json
// @Id      getSystemBackupFile
// @Param   fileID path string true "file setting ID"
// @Success 200 {object} filedto.GetFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/backup/files/{fileID} [get]
func (h *Handler) GetBackupFile(ctx *gin.Context) {
	fileID, err := h.ParseStringParam(ctx, "fileID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemBackup,
		ResourceID:     fileID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewGetFileReq()
	req.ID = fileID
	req.Types = []base.FileType{base.FileTypeSystemBackup}
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.FileUC.GetFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DownloadBackupFile Downloads a backup file
// @Summary Downloads a backup file
// @Description Downloads a backup file
// @Tags    system_settings
// @Produce json
// @Id      downloadSystemBackupFile
// @Param   fileID path string true "file setting ID"
// @Success 200
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/backup/files/{fileID}/download [get]
func (h *Handler) DownloadBackupFile(ctx *gin.Context) {
	fileID, err := h.ParseStringParam(ctx, "fileID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemBackup,
		ResourceID:     fileID,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewDownloadFileReq()
	req.ID = fileID
	req.UsePresignURLOnFileSize = defaultUsePresignURLOnFileSize
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.FileUC.DownloadFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	data := resp.Data

	if data.RedirectURL != "" {
		ctx.Redirect(http.StatusFound, data.RedirectURL)
		return
	}

	ctx.DataFromReader(http.StatusOK, data.ContentLength, data.ContentType, data.Content, data.ExtraHeaders)
}
