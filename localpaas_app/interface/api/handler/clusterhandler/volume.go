package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
)

// ListVolume Lists cluster volume settings
// @Summary Lists cluster volume settings
// @Description Lists cluster volume settings
// @Tags    cluster_volumes
// @Produce json
// @Id      listClusterVolume
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} volumedto.ListVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes [get]
func (h *Handler) ListVolume(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeGlobal)
}

// GetVolume Gets volume setting details
// @Summary Gets volume setting details
// @Description Gets volume setting details
// @Tags    cluster_volumes
// @Produce json
// @Id      getClusterVolume
// @Param   itemID path string true "setting ID"
// @Success 200 {object} volumedto.GetVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/{itemID} [get]
func (h *Handler) GetVolume(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeGlobal)
}

// CreateVolume Creates a new volume setting
// @Summary Creates a new volume setting
// @Description Creates a new volume setting
// @Tags    cluster_volumes
// @Produce json
// @Id      createClusterVolume
// @Param   body body volumedto.CreateVolumeReq true "request data"
// @Success 201 {object} volumedto.CreateVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes [post]
func (h *Handler) CreateVolume(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeGlobal)
}

// UpdateVolume Updates cluster volume
// @Summary Updates cluster volume
// @Description Updates cluster volume
// @Tags    cluster_volumes
// @Produce json
// @Id      updateClusterVolume
// @Param   itemID path string true "setting ID"
// @Param   body body volumedto.UpdateVolumeReq true "request data"
// @Success 200 {object} volumedto.UpdateVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/{itemID} [put]
func (h *Handler) UpdateVolume(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeGlobal)
}

// UpdateVolumeStatus Updates cluster volume status
// @Summary Updates cluster volume status
// @Description Updates cluster volume status
// @Tags    cluster_volumes
// @Produce json
// @Id      updateClusterVolumeStatus
// @Param   itemID path string true "setting ID"
// @Param   body body volumedto.UpdateVolumeStatusReq true "request data"
// @Success 200 {object} volumedto.UpdateVolumeStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/{itemID}/status [put]
func (h *Handler) UpdateVolumeStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeGlobal)
}

// DeleteVolume Deletes volume setting
// @Summary Deletes volume setting
// @Description Deletes volume setting
// @Tags    cluster_volumes
// @Produce json
// @Id      deleteClusterVolume
// @Param   itemID path string true "setting ID"
// @Success 200 {object} volumedto.DeleteVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/{itemID} [delete]
func (h *Handler) DeleteVolume(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeClusterVolume, base.ObjectScopeGlobal)
}

// SyncVolume Sync volumes from Docker
// @Summary Sync volumes from Docker
// @Description Sync volumes from Docker
// @Tags    cluster_volumes
// @Produce json
// @Id      syncClusterVolume
// @Param   body body volumedto.SyncVolumeReq true "request data"
// @Success 201 {object} volumedto.SyncVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/sync [post]
func (h *Handler) SyncVolume(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeClusterVolume, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewSyncVolumeReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ClusterVolumeUC.SyncVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
