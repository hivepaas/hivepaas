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
// @Tags    System settings
// @Produce json
// @Id      getSystemLoggingSettings
// @Param   revealSecrets query bool false "`revealSecrets=true` to include the secrets' values"
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
// @Tags    System settings
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

// GetLoggingPerformance Gets which nodes collect apps' routes and calls
// @Summary Gets which nodes collect apps' routes and calls
// @Description Lists the swarm's nodes for the collection of apps' routes and calls by OBI (eBPF): whether
// @Description each runs it, at which capacity - how many requests and connections OBI tracks at once, and so
// @Description its memory: small, medium or large, auto for the one HivePaaS recommends for the node's memory
// @Description - and what its agent last said of it: whether OBI runs, for how many apps, and whether the node
// @Description can run it, with why not - said every minute while the feature is on, every 10 minutes while it
// @Description is off. `statusReason` says why the statuses were not read: logs-not-stored, or unreadable.
// @Tags    System settings
// @Produce json
// @Id      getSystemLoggingPerformance
// @Success 200 {object} loggingdto.GetLoggingPerformanceResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/logging/performance [get]
func (h *Handler) GetLoggingPerformance(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeLogging,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := loggingdto.NewGetLoggingPerformanceReq()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.LoggingUC.GetLoggingPerformance(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateLoggingPerformance Updates which nodes collect apps' routes and calls
// @Summary Updates which nodes collect apps' routes and calls
// @Description Saves whether apps' routes and calls are collected by OBI (eBPF), and on which nodes, each with
// @Description its capacity. It is saved with the logging settings, whose `updateVer` it takes and changes,
// @Description and applies nothing: each node's agent reads it within 30 seconds. An app's routes and calls
// @Description are collected while both it and its node are on.
// @Tags    System settings
// @Produce json
// @Id      updateSystemLoggingPerformance
// @Param   body body loggingdto.UpdateLoggingPerformanceReq true "request data"
// @Success 200 {object} loggingdto.UpdateLoggingPerformanceResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/logging/performance [put]
func (h *Handler) UpdateLoggingPerformance(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeLogging,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := loggingdto.NewUpdateLoggingPerformanceReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.LoggingUC.UpdateLoggingPerformance(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
