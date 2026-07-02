package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
)

// ListNetwork Lists cluster network settings
// @Summary Lists cluster network settings
// @Description Lists cluster network settings
// @Tags    cluster_networks
// @Produce json
// @Id      listClusterNetwork
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} networkdto.ListNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks [get]
func (h *Handler) ListNetwork(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeGlobal)
}

// GetNetwork Gets network setting details
// @Summary Gets network setting details
// @Description Gets network setting details
// @Tags    cluster_networks
// @Produce json
// @Id      getClusterNetwork
// @Param   itemID path string true "setting ID"
// @Success 200 {object} networkdto.GetNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/{itemID} [get]
func (h *Handler) GetNetwork(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeGlobal)
}

// CreateNetwork Creates a new network setting
// @Summary Creates a new network setting
// @Description Creates a new network setting
// @Tags    cluster_networks
// @Produce json
// @Id      createClusterNetwork
// @Param   body body networkdto.CreateNetworkReq true "request data"
// @Success 201 {object} networkdto.CreateNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks [post]
func (h *Handler) CreateNetwork(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeGlobal)
}

// UpdateNetwork Updates cluster network
// @Summary Updates cluster network
// @Description Updates cluster network
// @Tags    cluster_networks
// @Produce json
// @Id      updateClusterNetwork
// @Param   itemID path string true "setting ID"
// @Param   body body networkdto.UpdateNetworkReq true "request data"
// @Success 200 {object} networkdto.UpdateNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/{itemID} [put]
func (h *Handler) UpdateNetwork(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeGlobal)
}

// UpdateNetworkStatus Updates cluster network status
// @Summary Updates cluster network status
// @Description Updates cluster network status
// @Tags    cluster_networks
// @Produce json
// @Id      updateClusterNetworkStatus
// @Param   itemID path string true "setting ID"
// @Param   body body networkdto.UpdateNetworkStatusReq true "request data"
// @Success 200 {object} networkdto.UpdateNetworkStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/{itemID}/status [put]
func (h *Handler) UpdateNetworkStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeGlobal)
}

// DeleteNetwork Deletes network setting
// @Summary Deletes network setting
// @Description Deletes network setting
// @Tags    cluster_networks
// @Produce json
// @Id      deleteClusterNetwork
// @Param   itemID path string true "setting ID"
// @Success 200 {object} networkdto.DeleteNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/{itemID} [delete]
func (h *Handler) DeleteNetwork(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeClusterNetwork, base.ObjectScopeGlobal)
}

// SyncNetwork Sync networks from Docker
// @Summary Sync networks from Docker
// @Description Sync networks from Docker
// @Tags    cluster_networks
// @Produce json
// @Id      syncClusterNetwork
// @Param   body body networkdto.SyncNetworkReq true "request data"
// @Success 201 {object} networkdto.SyncNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/sync [post]
func (h *Handler) SyncNetwork(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeClusterNetwork, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewSyncNetworkReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ClusterNetworkUC.SyncNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
