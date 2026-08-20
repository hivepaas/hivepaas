package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

func (h *Handler) GetAuth(
	ctx *gin.Context,
	action base.ActionType,
) (auth *basedto.Auth, projectID, projectEnvID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
	projectEnv, err := h.ParseStringParam(ctx, "projectEnv")
	if err != nil {
		return
	}
	projectEnvID = projecthelper.CalcProjectEnvID(projectID, projectEnv)
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		ProjectEnv:      &projectEnv,
	})
	if err != nil {
		return
	}
	return
}

func (h *Handler) GetAuthForItem(
	ctx *gin.Context,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, projectID, projectEnvID, itemID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
	projectEnv, err := h.ParseStringParam(ctx, "projectEnv")
	if err != nil {
		return
	}
	projectEnvID = projecthelper.CalcProjectEnvID(projectID, projectEnv)
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		ProjectEnv:      &projectEnv,
	})
	if err != nil {
		return
	}
	return
}
