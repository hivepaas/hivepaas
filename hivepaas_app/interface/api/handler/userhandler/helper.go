package userhandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

//nolint:nakedret
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
	var accessCheck permission.AccessCheck
	accessCheck = &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		Module:          base.ResourceModuleUser,
		ResourceType:    resType,
		ResourceID:      userID,
	}
	// Users read their own account. Changing or deleting it takes the Users
	// module, as anyone else's does: its role, its grants, its expiry are what an
	// admin decides about them, and their own profile and password have routes of
	// their own.
	ownRead := action == base.ActionTypeRead
	if userID == "current" && ownRead {
		accessCheck = authhandler.NoAccessCheck
	}
	auth, err = h.authHandler.GetCurrentAuth(ctx, accessCheck)
	if auth != nil && (userID == "current" || userID == auth.User.ID) {
		userID = auth.User.ID
		if ownRead {
			err = nil
		}
	}
	if err != nil {
		return
	}
	return
}
