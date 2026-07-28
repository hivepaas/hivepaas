package appbasehandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

//nolint:nakedret
func (h *Handler) GetAuth(
	ctx *gin.Context,
	action base.ActionType,
) (auth *basedto.Auth, projectID, projectEnvID, appID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
	projectEnv, err := h.ParseStringParam(ctx, "projectEnv")
	if err != nil {
		return
	}
	projectEnvID = projecthelper.CalcProjectEnvID(projectID, projectEnv)
	appID, err = h.ParseStringParam(ctx, "appID")
	if err != nil {
		return
	}
	accessCheck := &permission.AppAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		AppID:           appID,
		ProjectEnv:      projectEnvID,
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}

//nolint:nakedret
func (h *Handler) GetAuthInProject(
	ctx *gin.Context,
	action base.ActionType,
) (auth *basedto.Auth, projectID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
	accessCheck := &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}

//nolint:nakedret
func (h *Handler) GetAuthInEnv(
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
	accessCheck := &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		ProjectEnv:      &projectEnvID,
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}

//nolint:nakedret,unparam
func (h *Handler) GetAuthForItem(
	ctx *gin.Context,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, projectID, projectEnvID, appID, itemID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
	projectEnv, err := h.ParseStringParam(ctx, "projectEnv")
	if err != nil {
		return
	}
	projectEnvID = projecthelper.CalcProjectEnvID(projectID, projectEnv)
	appID, err = h.ParseStringParam(ctx, "appID")
	if err != nil {
		return
	}
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, &permission.AppAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		AppID:           appID,
		ProjectEnv:      projectEnvID,
	})
	if err != nil {
		return
	}
	return
}
