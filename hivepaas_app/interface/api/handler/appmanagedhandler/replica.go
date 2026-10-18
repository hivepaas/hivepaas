package appmanagedhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appmanageduc/appmanageddto"
)

// AddManagedReplica Adds a read replica to a managed database
// @Summary Adds a read replica to a managed database
// @Description An asynchronous replica on a volume of its own, pinned to a node: seeded from the database's
// @Description archive, which must be in object storage, then streaming from the database, which is told to
// @Description let it in first. The database's scripts must be ones that can run a replica.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedReplicaAdd
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appmanageddto.AddManagedReplicaReq true "request data"
// @Success 201 {object} appmanageddto.AddManagedReplicaResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/replicas [post]
func (h *Handler) AddManagedReplica(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewAddManagedReplicaReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.AddManagedReplica(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, resp)
}

// RemoveManagedReplica Removes a read replica from a managed database
// @Summary Removes a read replica from a managed database
// @Description The replica's app is deleted, with its directory of its volume.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedReplicaRemove
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   memberID path string true "the replica's app ID"
// @Success 200 {object} appmanageddto.RemoveManagedReplicaResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/replicas/{memberID} [delete]
func (h *Handler) RemoveManagedReplica(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewRemoveManagedReplicaReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	req.MemberID = ctx.Param("memberID")
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.RemoveManagedReplica(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// PromoteManagedMember Promotes a replica to a managed database's primary
// @Summary Promotes a replica to a managed database's primary
// @Description In order: what writes now is fenced - kept from starting its server, its service stopped, and
// @Description seen stopped; the replica is promoted; the engine's jobs move to it, and the other replicas are
// @Description restarted to follow it. Before the replica is promoted, a failure lifts the fence; once whether it
// @Description was is not known, the fence stays, and promoting it again finishes the promotion.
// @Tags    Managed apps
// @Produce json
// @Id      appManagedPromote
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body appmanageddto.PromoteManagedMemberReq true "request data"
// @Success 200 {object} appmanageddto.PromoteManagedMemberResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 412 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/managed/promote [post]
func (h *Handler) PromoteManagedMember(ctx *gin.Context) {
	auth, projectID, projectEnvID, appID, err := h.GetAuth(ctx, base.ActionTypeWrite)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := appmanageddto.NewPromoteManagedMemberReq()
	req.ProjectID = projectID
	req.ProjectEnvID = projectEnvID
	req.AppID = appID
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.appManagedUC.PromoteManagedMember(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
