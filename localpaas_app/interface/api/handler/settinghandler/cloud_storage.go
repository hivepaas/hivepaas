package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/cloudstorageuc/cloudstoragedto"
)

// ListCloudStorage Lists cloud storages
// @Summary Lists cloud storages
// @Description Lists cloud storages
// @Tags    settings
// @Produce json
// @Id      listSettingCloudStorage
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} cloudstoragedto.ListCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/cloud-storages [get]
func (h *Handler) ListCloudStorage(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeGlobal)
}

// GetCloudStorage Gets cloud storage details
// @Summary Gets cloud storage details
// @Description Gets cloud storage details
// @Tags    settings
// @Produce json
// @Id      getSettingCloudStorage
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cloudstoragedto.GetCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/cloud-storages/{itemID} [get]
func (h *Handler) GetCloudStorage(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeGlobal)
}

// CreateCloudStorage Creates a new cloud storage
// @Summary Creates a new cloud storage
// @Description Creates a new cloud storage
// @Tags    settings
// @Produce json
// @Id      createSettingCloudStorage
// @Param   body body cloudstoragedto.CreateCloudStorageReq true "request data"
// @Success 201 {object} cloudstoragedto.CreateCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/cloud-storages [post]
func (h *Handler) CreateCloudStorage(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeGlobal)
}

// UpdateCloudStorage Updates a cloud storage
// @Summary Updates a cloud storage
// @Description Updates a cloud storage
// @Tags    settings
// @Produce json
// @Id      updateSettingCloudStorage
// @Param   itemID path string true "setting ID"
// @Param   body body cloudstoragedto.UpdateCloudStorageReq true "request data"
// @Success 200 {object} cloudstoragedto.UpdateCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/cloud-storages/{itemID} [put]
func (h *Handler) UpdateCloudStorage(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeGlobal)
}

// UpdateCloudStorageStatus Updates a cloud storage's status
// @Summary Updates a cloud storage's status
// @Description Updates a cloud storage's status
// @Tags    settings
// @Produce json
// @Id      updateSettingCloudStorageStatus
// @Param   itemID path string true "setting ID"
// @Param   body body cloudstoragedto.UpdateCloudStorageStatusReq true "request data"
// @Success 200 {object} cloudstoragedto.UpdateCloudStorageStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/cloud-storages/{itemID}/status [put]
func (h *Handler) UpdateCloudStorageStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeGlobal)
}

// DeleteCloudStorage Deletes a cloud storage
// @Summary Deletes a cloud storage
// @Description Deletes a cloud storage
// @Tags    settings
// @Produce json
// @Id      deleteSettingCloudStorage
// @Param   itemID path string true "setting ID"
// @Success 200 {object} cloudstoragedto.DeleteCloudStorageResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/cloud-storages/{itemID} [delete]
func (h *Handler) DeleteCloudStorage(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeCloudStorage, base.ObjectScopeGlobal)
}

// TestCloudStorageConn Test cloud storage connection
// @Summary Test cloud storage connection
// @Description Test cloud storage connection
// @Tags    settings
// @Produce json
// @Id      testCloudStorageConn
// @Param   body body cloudstoragedto.TestCloudStorageConnReq true "request data"
// @Success 200 {object} cloudstoragedto.TestCloudStorageConnResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/cloud-storages/test-conn [post]
func (h *Handler) TestCloudStorageConn(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := cloudstoragedto.NewTestCloudStorageConnReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.CloudStorageUC.TestCloudStorageConn(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
