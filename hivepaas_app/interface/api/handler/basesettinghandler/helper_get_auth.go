package basesettinghandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
)

func (h *Handler) GetAuthGlobalSettings(
	ctx *gin.Context,
	resourceType base.ResourceType,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, itemID string, err error) {
	return h.GetAuthGlobalSettingsAnyAction(ctx, resourceType, []base.ActionType{action}, paramName)
}

func (h *Handler) GetAuthGlobalSettingsAnyAction(
	ctx *gin.Context,
	resourceType base.ResourceType,
	anyActions []base.ActionType,
	paramName string,
) (auth *basedto.Auth, itemID string, err error) {
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}

	accessCheck := &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{},
		Module:          base.ResourceModuleSettings,
		ResourceType:    resourceType,
		ResourceID:      itemID,
	}
	if len(anyActions) == 1 {
		accessCheck.Action = anyActions[0]
	} else {
		accessCheck.AnyOf = anyActions
	}

	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}

func (h *Handler) GetAuthUserSettings(
	ctx *gin.Context,
	_ base.ActionType,
	paramName string,
) (auth *basedto.Auth, userID, itemID string, err error) {
	auth, err = h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		return
	}
	userID = auth.User.ID
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	return
}

func (h *Handler) GetAuthProjectSettings(
	ctx *gin.Context,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, projectID, itemID string, err error) {
	return h.GetAuthProjectSettingsAnyAction(ctx, []base.ActionType{action}, paramName)
}

//nolint:nakedret
func (h *Handler) GetAuthProjectSettingsAnyAction(
	ctx *gin.Context,
	anyActions []base.ActionType,
	paramName string,
) (auth *basedto.Auth, projectID, itemID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}

	accessCheck := &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{},
		ProjectID:       projectID,
	}
	if len(anyActions) == 1 {
		accessCheck.Action = anyActions[0]
	} else {
		accessCheck.AnyOf = anyActions
	}

	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}

func (h *Handler) GetAuthProjectEnvSettings(
	ctx *gin.Context,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, projectID, projectEnvID, itemID string, err error) {
	return h.GetAuthProjectEnvSettingsAnyAction(ctx, []base.ActionType{action}, paramName)
}

//nolint:nakedret
func (h *Handler) GetAuthProjectEnvSettingsAnyAction(
	ctx *gin.Context,
	anyActions []base.ActionType,
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

	accessCheck := &permission.ProjectAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{},
		ProjectID:       projectID,
		ProjectEnv:      &projectEnvID,
	}
	if len(anyActions) == 1 {
		accessCheck.Action = anyActions[0]
	} else {
		accessCheck.AnyOf = anyActions
	}

	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}

func (h *Handler) GetAuthAppSettings(
	ctx *gin.Context,
	action base.ActionType,
	paramName string,
) (auth *basedto.Auth, projectID, projectEnvID, appID, itemID string, err error) {
	return h.GetAuthAppSettingsAnyAction(ctx, []base.ActionType{action}, paramName)
}

//nolint:nakedret
func (h *Handler) GetAuthAppSettingsAnyAction(
	ctx *gin.Context,
	anyActions []base.ActionType,
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

	accessCheck := &permission.AppAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{},
		AppID:           appID,
		ProjectID:       projectID,
		ProjectEnv:      projectEnvID,
	}
	if len(anyActions) == 1 {
		accessCheck.Action = anyActions[0]
	} else {
		accessCheck.AnyOf = anyActions
	}

	auth, err = h.AuthHandler.GetCurrentAuth(ctx, accessCheck)
	if err != nil {
		return
	}
	return
}
