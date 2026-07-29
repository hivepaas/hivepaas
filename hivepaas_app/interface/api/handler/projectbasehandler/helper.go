package projectbasehandler

import (
	"github.com/gin-gonic/gin"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
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
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
	})
	if err != nil {
		return
	}
	return
}

func (h *Handler) GetEnvAuth(
	ctx *gin.Context,
	action base.ActionType,
	getEnvID bool,
) (auth *basedto.Auth, projectID, projectEnvID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
	if getEnvID {
		var projectEnv string
		projectEnv, err = h.ParseStringParam(ctx, "projectEnv")
		if err != nil {
			return
		}
		projectEnvID = projecthelper.CalcProjectEnvID(projectID, projectEnv)
	}
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: action},
		ProjectID:       projectID,
		ProjectEnv:      gofn.If(getEnvID, &projectEnvID, nil),
	})
	if err != nil {
		return
	}
	return
}
