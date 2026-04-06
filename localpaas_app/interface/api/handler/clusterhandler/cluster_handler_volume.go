package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
)

// ListVolume Lists cluster volumes
// @Summary Lists cluster volumes
// @Description Lists cluster volumes
// @Tags    cluster_volumes
// @Produce json
// @Id      listClusterVolume
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} volumedto.ListVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes [get]
func (h *ClusterHandler) ListVolume(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeVolume, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewListVolumeReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.volumeUC.ListVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetVolume Gets volume details
// @Summary Gets volume details
// @Description Gets volume details
// @Tags    cluster_volumes
// @Produce json
// @Id      getClusterVolume
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.GetVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/{volumeID} [get]
func (h *ClusterHandler) GetVolume(ctx *gin.Context) {
	auth, volumeID, err := h.getAuth(ctx, base.ResourceTypeVolume, base.ActionTypeRead, "volumeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewGetVolumeReq()
	req.VolumeID = volumeID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.volumeUC.GetVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetVolumeInspection Gets volume details
// @Summary Gets volume details
// @Description Gets volume details
// @Tags    cluster_volumes
// @Produce json
// @Id      getClusterVolumeInspection
// @Param   volumeID path string true "volume ID"
// @Success 200 {object} volumedto.GetVolumeInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/{volumeID}/inspect [get]
func (h *ClusterHandler) GetVolumeInspection(ctx *gin.Context) {
	auth, volumeID, err := h.getAuth(ctx, base.ResourceTypeVolume, base.ActionTypeRead, "volumeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewGetVolumeInspectionReq()
	req.VolumeID = volumeID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.volumeUC.GetVolumeInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateVolume Creates a volume
// @Summary Creates a volume
// @Description Creates a volume
// @Tags    cluster_volumes
// @Produce json
// @Id      createClusterVolume
// @Param   body body volumedto.CreateVolumeReq true "request data"
// @Success 200 {object} volumedto.CreateVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes [post]
func (h *ClusterHandler) CreateVolume(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeVolume, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewCreateVolumeReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.volumeUC.CreateVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteVolume Deletes a volume
// @Summary Deletes a volume
// @Description Deletes a volume
// @Tags    cluster_volumes
// @Produce json
// @Id      deleteClusterVolume
// @Param   volumeID path string true "volume ID"
// @Param   force query bool false "`force=true/false`"
// @Success 200 {object} volumedto.DeleteVolumeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/volumes/{volumeID} [delete]
func (h *ClusterHandler) DeleteVolume(ctx *gin.Context) {
	auth, volumeID, err := h.getAuth(ctx, base.ResourceTypeVolume, base.ActionTypeDelete, "volumeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := volumedto.NewDeleteVolumeReq()
	req.VolumeID = volumeID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.volumeUC.DeleteVolume(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
