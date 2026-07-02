package userhandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

func (h *Handler) getAuth(
	ctx *gin.Context,
	resType base.ResourceType,
	action base.ActionType,
	getUserID bool,
) (auth *basedto.Auth, userID string, err error) {
	if getUserID {
		userID, err = h.ParseStringParam(ctx, "userID")
		if err != nil {
			return
		}
	}
	accessCheck := &permission.AccessCheck{
		ResourceModule: base.ResourceModuleUser,
		ResourceType:   resType,
		ResourceID:     userID,
		Action:         action,
	}
	if userID == "current" {
		accessCheck = authhandler.NoAccessCheck
	}
	auth, err = h.authHandler.GetCurrentAuth(ctx, accessCheck)
	if auth != nil && (userID == "current" || userID == auth.User.ID) {
		err = nil
		userID = auth.User.ID
	}
	if err != nil {
		return
	}
	return
}
