package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

// GetBackupSettings Gets backup settings
// @Summary Gets backup settings
// @Description Gets backup settings
// @Tags    system_settings
// @Produce json
// @Id      getSystemBackupSettings
// @Success 200 {object} systembackupdto.GetSystemBackupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/backup [get]
func (h *Handler) GetBackupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemBackup,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systembackupdto.NewGetSystemBackupReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SystemBackupUC.GetSystemBackup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateBackupSettings Updates backup settings
// @Summary Updates backup settings
// @Description Updates backup settings
// @Tags    system_settings
// @Produce json
// @Id      updateSystemBackupSettings
// @Param   body body systembackupdto.UpdateSystemBackupReq true "request data"
// @Success 200 {object} systembackupdto.UpdateSystemBackupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/backup [put]
func (h *Handler) UpdateBackupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemBackup,
		Action:         base.ActionTypeWrite,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systembackupdto.NewUpdateSystemBackupReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SystemBackupUC.UpdateSystemBackup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ExecuteBackup Executes the backup
// @Summary Executes the backup
// @Description Executes the backup
// @Tags    system_settings
// @Produce json
// @Id      executeSystemBackup
// @Param   body body systembackupdto.ExecuteSystemBackupReq true "request data"
// @Success 200 {object} systembackupdto.ExecuteSystemBackupResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /system/settings/backup/exec [post]
func (h *Handler) ExecuteBackup(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSystem,
		ResourceType:   base.ResourceTypeSystemBackup,
		Action:         base.ActionTypeExecute,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systembackupdto.NewExecuteSystemBackupReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SystemBackupUC.ExecuteSystemBackup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
