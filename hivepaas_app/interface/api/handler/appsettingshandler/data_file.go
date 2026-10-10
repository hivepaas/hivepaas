package appsettingshandler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

// CreateDataFile Creates a data file of an app
// @Summary Creates a data file of an app
// @Description Creates a data file of an app
// @Tags    App settings
// @Produce json
// @Id      createAppDataFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body filedto.CreateFileReq true "request data"
// @Success 201 {object} filedto.CreateFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/data-files [post]
func (h *Handler) CreateDataFile(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, _, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewCreateFileReq()
	req.Scope = entity.NewObjectScopeApp(appID, "", projectID, projectEnvID)
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.CreateFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// ListDataFile Lists data files of an app
// @Summary Lists data files of an app
// @Description Lists data files of an app
// @Tags    App settings
// @Produce json
// @Id      listAppDataFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   key query string false "`key=<key>`, comma separated"
// @Param   kind query string false "`kind=<file kind>`, comma separated"
// @Param   search query string false "`search=<text>`"
// @Param   status query string false "`status=<status>`, comma separated"
// @Param   storageType query string false "`storageType=volume` or `cloud`, comma separated"
// @Param   type query string false "`type=<file type>`, comma separated"
// @Success 200 {object} filedto.ListFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/data-files [get]
func (h *Handler) ListDataFile(ctx *gin.Context) {
	auth, _, _, appID, err := h.GetAuth(ctx, base.ActionTypeRead)
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
// @Tags    App settings
// @Produce json
// @Id      getAppDataFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "file ID"
// @Param   kind query string false "`kind=<file kind>`, comma separated: the file is one of them"
// @Param   type query string false "`type=<file type>`, comma separated: the file is one of them"
// @Success 200 {object} filedto.GetFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/data-files/{itemID} [get]
func (h *Handler) GetDataFile(ctx *gin.Context) {
	auth, _, _, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "itemID")
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
// @Tags    App settings
// @Produce json
// @Id      getAppDataFileDownloadURL
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "file ID"
// @Param   viewInline query bool false "`viewInline=true` for a URL the browser shows rather than saves"
// @Success 200 {object} filedto.GetFileDownloadURLResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/data-files/{itemID}/download-url [get]
func (h *Handler) GetDataFileDownloadURL(ctx *gin.Context) {
	auth, _, _, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeRead, "itemID")
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
// @Tags    App settings
// @Produce json
// @Id      deleteAppDataFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "file ID"
// @Param   deletePermanently query bool false "`deletePermanently=true` also deletes the file's data, wherever it is"
// @Param   deletePermanentlyIfOnVolume query bool false "`=true` deletes the data too if it is on a volume"
// @Success 200 {object} filedto.DeleteFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/data-files/{itemID} [delete]
func (h *Handler) DeleteDataFile(ctx *gin.Context) {
	auth, _, _, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "itemID")
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

// LoadDataFile Loads a data file into a command
// @Summary Loads a data file into a command
// @Description Feeds an app's data file to a command run in the app, on its stdin - a dump a job
// @Description saved, loaded back into the app's database. A file saved encrypted takes its passphrase.
// @Description The load runs as a task, whose ID is answered.
// @Tags    App settings
// @Produce json
// @Id      loadAppDataFile
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "file ID"
// @Param   body body filedto.LoadDataFileReq true "request data"
// @Success 200 {object} filedto.LoadDataFileResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/data-files/{itemID}/load [post]
func (h *Handler) LoadDataFile(ctx *gin.Context) {
	// What it loads changes the app's data, as a restore does.
	auth, projectID, _, appID, itemID, err := h.GetAuthForItem(ctx, base.ActionTypeWrite, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := filedto.NewLoadDataFileReq()
	req.ID = itemID
	req.ProjectID = projectID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.LoadDataFile(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
