package filehandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/fileuc/filedto"
)

// UploadFiles Uploads one or multiple files to server
// @Summary Uploads one or multiple files to server
// @Description Uploads one or multiple files to server
// @Tags    files
// @Accept  multipart/form-data
// @Produce json
// @Id      uploadFiles
// @Param   file formData file true "one or multiple files to upload"
// @Param   type formData string true "file type: `build-source,...`"
// @Param   scope formData string true "object target scope: project/app/user/global"
// @Param   projectId formData string false "target project id if scope=project or app"
// @Param   appId formData string false "target app id if scope=app"
// @Param   userId formData string false "target app id if scope=user"
// @Success 200 {object} filedto.UploadResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /files/upload [post]
func (h *Handler) UploadFiles(ctx *gin.Context) {
	req := filedto.NewUploadReq()
	err := h.ParseFormFiles(ctx, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	auth, err := h.checkUploadPermission(ctx, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.fileUC.Upload(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) checkUploadPermission(ctx *gin.Context, req *filedto.UploadReq) (auth *basedto.Auth, err error) {
	var accessCheck *permission.AccessCheck
	switch req.FileType {
	case base.FileTypeBuildSource:
		accessCheck = &permission.AccessCheck{
			ResourceModule:     base.ResourceModuleProject,
			ResourceType:       base.ResourceTypeApp,
			ResourceID:         req.Scope.AppID,
			ParentResourceType: base.ResourceTypeProject,
			ParentResourceID:   req.Scope.ProjectID,
			AnyOf:              []base.ActionType{base.ActionTypeExecute, base.ActionTypeWrite},
		}
	case base.FileTypeSystemBackup, base.FileTypeRepoCache:
		fallthrough
	case base.FileTypeSchedJobOutput:
		fallthrough
	default:
		return nil, apperrors.New(apperrors.ErrFileTypeNotSupported).
			WithParam("SupportedTypes", []base.FileType{base.FileTypeBuildSource})
	}

	auth, err = h.authHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return auth, nil
}
