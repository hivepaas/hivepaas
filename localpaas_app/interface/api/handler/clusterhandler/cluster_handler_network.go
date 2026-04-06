package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/networkuc/networkdto"
)

// ListNetwork Lists cluster networks
// @Summary Lists cluster networks
// @Description Lists cluster networks
// @Tags    cluster_networks
// @Produce json
// @Id      listClusterNetwork
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} networkdto.ListNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks [get]
func (h *ClusterHandler) ListNetwork(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeNetwork, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewListNetworkReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.networkUC.ListNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetNetwork Gets network details
// @Summary Gets network details
// @Description Gets network details
// @Tags    cluster_networks
// @Produce json
// @Id      getClusterNetwork
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.GetNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/{networkID} [get]
func (h *ClusterHandler) GetNetwork(ctx *gin.Context) {
	auth, networkID, err := h.getAuth(ctx, base.ResourceTypeNetwork, base.ActionTypeRead, "networkID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewGetNetworkReq()
	req.NetworkID = networkID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.networkUC.GetNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetNetworkInspection Gets network details
// @Summary Gets network details
// @Description Gets network details
// @Tags    cluster_networks
// @Produce json
// @Id      getClusterNetworkInspection
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.GetNetworkInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/{networkID}/inspect [get]
func (h *ClusterHandler) GetNetworkInspection(ctx *gin.Context) {
	auth, networkID, err := h.getAuth(ctx, base.ResourceTypeNetwork, base.ActionTypeRead, "networkID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewGetNetworkInspectionReq()
	req.NetworkID = networkID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.networkUC.GetNetworkInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// CreateNetwork Creates a network
// @Summary Creates a network
// @Description Creates a network
// @Tags    cluster_networks
// @Produce json
// @Id      createClusterNetwork
// @Param   body body networkdto.CreateNetworkReq true "request data"
// @Success 200 {object} networkdto.CreateNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks [post]
func (h *ClusterHandler) CreateNetwork(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeNetwork, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewCreateNetworkReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.networkUC.CreateNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// DeleteNetwork Deletes a network
// @Summary Deletes a network
// @Description Deletes a network
// @Tags    cluster_networks
// @Produce json
// @Id      deleteClusterNetwork
// @Param   networkID path string true "network ID"
// @Success 200 {object} networkdto.DeleteNetworkResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/networks/{networkID} [delete]
func (h *ClusterHandler) DeleteNetwork(ctx *gin.Context) {
	auth, networkID, err := h.getAuth(ctx, base.ResourceTypeNetwork, base.ActionTypeDelete, "networkID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := networkdto.NewDeleteNetworkReq()
	req.NetworkID = networkID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.networkUC.DeleteNetwork(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
