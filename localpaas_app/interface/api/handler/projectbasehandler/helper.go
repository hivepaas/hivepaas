package projectbasehandler

import (
	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/permission"
)

func (h *Handler) GetAuth(
	ctx *gin.Context,
	action base.ActionType,
	getProjectID bool,
) (auth *basedto.Auth, projectID string, err error) {
	if getProjectID {
		projectID, err = h.ParseStringParam(ctx, "projectID")
		if err != nil {
			return
		}
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleProject,
		ResourceType:   base.ResourceTypeProject,
		ResourceID:     projectID,
		Action:         action,
	})
	if err != nil {
		return
	}
	return
}
