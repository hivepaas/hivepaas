package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/logginguc/loggingdto"
)

// GetLoggingSettings Gets logging settings
// @Summary Gets logging settings
// @Description Gets logging settings
// @Tags    system_settings
// @Produce json
// @Id      getSystemLoggingSettings
// @Success 200 {object} loggingdto.GetLoggingSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/logging [get]
func (h *Handler) GetLoggingSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeLogging,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := loggingdto.NewGetLoggingSettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.LoggingUC.GetLoggingSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateLoggingSettings Updates logging settings
// @Summary Updates logging settings
// @Description Updates logging settings
// @Tags    system_settings
// @Produce json
// @Id      updateSystemLoggingSettings
// @Param   body body loggingdto.UpdateLoggingSettingsReq true "request data"
// @Success 200 {object} loggingdto.UpdateLoggingSettingsResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/logging [put]
func (h *Handler) UpdateLoggingSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeLogging,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := loggingdto.NewUpdateLoggingSettingsReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.LoggingUC.UpdateLoggingSettings(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
