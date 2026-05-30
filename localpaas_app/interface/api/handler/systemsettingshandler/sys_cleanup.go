package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systemcleanupuc/systemcleanupdto"
)

// GetCleanupSettings Gets cleanup settings
// @Summary Gets cleanup settings
// @Description Gets cleanup settings
// @Tags    system_settings
// @Produce json
// @Id      getSystemCleanupSettings
// @Success 200 {object} systemcleanupdto.GetSystemCleanupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/cleanup [get]
func (h *Handler) GetCleanupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemCleanup,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systemcleanupdto.NewGetSystemCleanupReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SystemCleanupUC.GetSystemCleanup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateCleanupSettings Updates cleanup settings
// @Summary Updates cleanup settings
// @Description Updates cleanup settings
// @Tags    system_settings
// @Produce json
// @Id      updateSystemCleanupSettings
// @Param   body body systemcleanupdto.UpdateSystemCleanupReq true "request data"
// @Success 200 {object} systemcleanupdto.UpdateSystemCleanupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/cleanup [put]
func (h *Handler) UpdateCleanupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemCleanup,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systemcleanupdto.NewUpdateSystemCleanupReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SystemCleanupUC.UpdateSystemCleanup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ExecuteCleanup Executes the cleanup
// @Summary Executes the cleanup
// @Description Executes the cleanup
// @Tags    system_settings
// @Produce json
// @Id      executeSystemCleanup
// @Param   body body systemcleanupdto.ExecuteSystemCleanupReq true "request data"
// @Success 200 {object} systemcleanupdto.ExecuteSystemCleanupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/cleanup/exec [post]
func (h *Handler) ExecuteCleanup(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemCleanup,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systemcleanupdto.NewExecuteSystemCleanupReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SystemCleanupUC.ExecuteSystemCleanup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
