package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

// ListNode Lists cluster node settings
// @Summary Lists cluster node settings
// @Description Lists cluster node settings
// @Tags    cluster_nodes
// @Produce json
// @Id      listClusterNode
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} nodedto.ListNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes [get]
func (h *Handler) ListNode(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeClusterNode, base.ObjectScopeGlobal)
}

// GetNode Gets node setting details
// @Summary Gets node setting details
// @Description Gets node setting details
// @Tags    cluster_nodes
// @Produce json
// @Id      getClusterNode
// @Param   itemID path string true "setting ID"
// @Success 200 {object} nodedto.GetNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/{itemID} [get]
func (h *Handler) GetNode(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeClusterNode, base.ObjectScopeGlobal)
}

// UpdateNode Updates cluster node
// @Summary Updates cluster node
// @Description Updates cluster node
// @Tags    cluster_nodes
// @Produce json
// @Id      updateClusterNode
// @Param   itemID path string true "setting ID"
// @Param   body body nodedto.UpdateNodeReq true "request data"
// @Success 200 {object} nodedto.UpdateNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/{itemID} [put]
func (h *Handler) UpdateNode(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeClusterNode, base.ObjectScopeGlobal)
}

// DeleteNode Deletes node setting
// @Summary Deletes node setting
// @Description Deletes node setting
// @Tags    cluster_nodes
// @Produce json
// @Id      deleteClusterNode
// @Param   itemID path string true "setting ID"
// @Success 200 {object} nodedto.DeleteNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/{itemID} [delete]
func (h *Handler) DeleteNode(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeClusterNode, base.ObjectScopeGlobal)
}

// JoinNode Joins a node to the swarm
// @Summary Joins a node to the swarm
// @Description Joins a node to the swarm
// @Tags    cluster_nodes
// @Produce json
// @Id      joinClusterNode
// @Param   body body nodedto.JoinNodeReq true "request data"
// @Success 200 {object} nodedto.JoinNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/join [post]
func (h *Handler) JoinNode(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewJoinNodeReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.JoinNode(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetNodeJoinCommand Gets node join command
// @Summary Gets node join command
// @Description Gets node join command
// @Tags    cluster_nodes
// @Produce json
// @Id      getClusterNodeJoinCommand
// @Param   joinAsManager query string false "joinAsManager=true/false"
// @Success 200 {object} nodedto.GetNodeJoinCommandResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/join-command [get]
func (h *Handler) GetNodeJoinCommand(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewGetNodeJoinCommandReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.GetNodeJoinCommand(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// SetManagerNodes Sets manager nodes
// @Summary Sets manager nodes
// @Description Sets manager nodes
// @Tags    cluster_nodes
// @Produce json
// @Id      setClusterManagerNodes
// @Param   body body nodedto.SetManagerNodesReq true "request data"
// @Success 200 {object} nodedto.SetManagerNodesResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/set-managers [post]
func (h *Handler) SetManagerNodes(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeWrite, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewSetManagerNodesReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.SetManagerNodes(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
