package clusterhandler

import (
	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/permission"
)

func (h *Handler) getAuth(
	ctx *gin.Context,
	resType base.ResourceType,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, itemID string, err error) {
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleCluster,
		ResourceType:   resType,
		ResourceID:     itemID,
		Action:         action,
	})
	if err != nil {
		return
	}
	return
}
