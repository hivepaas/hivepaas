package basesettinghandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
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
	accessCheck := &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSettings,
		ResourceType:   resourceType,
		ResourceID:     itemID,
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
	accessCheck := &permission.AccessCheck{
		ResourceModule: base.ResourceModuleProject,
		ResourceType:   base.ResourceTypeProject,
		ResourceID:     projectID,
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
) (auth *basedto.Auth, projectID, appID, itemID string, err error) {
	return h.GetAuthAppSettingsAnyAction(ctx, []base.ActionType{action}, paramName)
}

//nolint:nakedret
func (h *Handler) GetAuthAppSettingsAnyAction(
	ctx *gin.Context,
	anyActions []base.ActionType,
	paramName string,
) (auth *basedto.Auth, projectID, appID, itemID string, err error) {
	projectID, err = h.ParseStringParam(ctx, "projectID")
	if err != nil {
		return
	}
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
	accessCheck := &permission.AccessCheck{
		ResourceModule:     base.ResourceModuleProject,
		ParentResourceType: base.ResourceTypeProject,
		ParentResourceID:   projectID,
		ResourceType:       base.ResourceTypeApp,
		ResourceID:         appID,
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
