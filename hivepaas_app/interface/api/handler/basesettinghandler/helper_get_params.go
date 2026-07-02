package basesettinghandler

import (
	"github.com/gin-gonic/gin"
)

func (h *Handler) GetParamGlobalSettings(
	ctx *gin.Context,
	paramName string,
) (itemID string, err error) {
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	return
}

func (h *Handler) GetParamUserSettings(
	ctx *gin.Context,
	paramName string,
) (itemID string, err error) {
	if paramName != "" {
		itemID, err = h.ParseStringParam(ctx, paramName)
		if err != nil {
			return
		}
	}
	return
}

func (h *Handler) GetParamProjectSettings(
	ctx *gin.Context,
	paramName string,
) (projectID, itemID string, err error) {
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
	return
}

func (h *Handler) GetParamAppSettings(
	ctx *gin.Context,
	paramName string,
) (projectID, appID, itemID string, err error) {
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
	return
}
