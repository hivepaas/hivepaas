package clusterhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/nodeuc/nodedto"
)

// ListNode Lists cluster nodes
// @Summary Lists cluster nodes
// @Description Lists cluster nodes
// @Tags    cluster_nodes
// @Produce json
// @Id      listClusterNode
// @Param   status query string false "`status=<target>`"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} nodedto.ListNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes [get]
func (h *Handler) ListNode(ctx *gin.Context) {
	auth, _, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewListNodeReq()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.ListNode(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetNode Gets node details
// @Summary Gets node details
// @Description Gets node details
// @Tags    cluster_nodes
// @Produce json
// @Id      getClusterNode
// @Param   nodeID path string true "node ID"
// @Success 200 {object} nodedto.GetNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/{nodeID} [get]
func (h *Handler) GetNode(ctx *gin.Context) {
	auth, nodeID, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeRead, "nodeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewGetNodeReq()
	req.NodeID = nodeID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.GetNode(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// GetNodeInspection Gets node details
// @Summary Gets node details
// @Description Gets node details
// @Tags    cluster_nodes
// @Produce json
// @Id      getClusterNodeInspection
// @Param   nodeID path string true "node ID"
// @Success 200 {object} nodedto.GetNodeInspectionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/{nodeID}/inspect [get]
func (h *Handler) GetNodeInspection(ctx *gin.Context) {
	auth, nodeID, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeRead, "nodeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewGetNodeInspectionReq()
	req.NodeID = nodeID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.GetNodeInspection(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateNode Updates a node
// @Summary Updates a node
// @Description Updates a node
// @Tags    cluster_nodes
// @Produce json
// @Id      updateClusterNode
// @Param   nodeID path string true "node ID"
// @Param   body body nodedto.UpdateNodeReq true "request data"
// @Success 200 {object} nodedto.UpdateNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/{nodeID} [put]
func (h *Handler) UpdateNode(ctx *gin.Context) {
	auth, nodeID, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeWrite, "nodeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewUpdateNodeReq()
	req.NodeID = nodeID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.UpdateNode(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// DeleteNode Deletes a node
// @Summary Deletes a node
// @Description Deletes a node
// @Tags    cluster_nodes
// @Produce json
// @Id      deleteClusterNode
// @Param   nodeID path string true "node ID"
// @Param   force query bool false "`force=true/false`"
// @Success 200 {object} nodedto.DeleteNodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /cluster/nodes/{nodeID} [delete]
func (h *Handler) DeleteNode(ctx *gin.Context) {
	auth, nodeID, err := h.getAuth(ctx, base.ResourceTypeNode, base.ActionTypeDelete, "nodeID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := nodedto.NewDeleteNodeReq()
	req.NodeID = nodeID
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil { // to make sure Validate() to be called
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.nodeUC.DeleteNode(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
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
