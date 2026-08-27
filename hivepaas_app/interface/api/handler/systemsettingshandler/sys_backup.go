package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/systembackupuc/systembackupdto"
)

// GetBackupSettings Gets backup settings
// @Summary Gets backup settings
// @Description Gets backup settings
// @Tags    system_settings
// @Produce json
// @Id      getSystemBackupSettings
// @Success 200 {object} systembackupdto.GetSystemBackupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/backup [get]
func (h *Handler) GetBackupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeSystemBackup,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systembackupdto.NewGetSystemBackupReq()
	req.Scope = entity.NewObjectScopeHivepaas()
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
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/backup [put]
func (h *Handler) UpdateBackupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeSystemBackup,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systembackupdto.NewUpdateSystemBackupReq()
	req.Scope = entity.NewObjectScopeHivepaas()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
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
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/backup/exec [post]
func (h *Handler) ExecuteBackup(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeExecute},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeSystemBackup,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := systembackupdto.NewExecuteSystemBackupReq()
	req.Scope = entity.NewObjectScopeHivepaas()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
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
