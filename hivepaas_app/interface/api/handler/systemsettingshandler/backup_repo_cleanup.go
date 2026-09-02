package systemsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/backuprepocleanupuc/backuprepocleanupdto"
)

// GetBackupRepoCleanupSettings Gets backup repo cleanup settings
// @Summary Gets backup repo cleanup settings
// @Description Gets backup repo cleanup settings
// @Tags    system_settings
// @Produce json
// @Id      getBackupRepoCleanupSettings
// @Success 200 {object} backuprepocleanupdto.GetBackupRepoCleanupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/backup-repo-cleanup [get]
func (h *Handler) GetBackupRepoCleanupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeRead},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeBackupRepoCleanup,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := backuprepocleanupdto.NewGetBackupRepoCleanupReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.BackupRepoCleanupUC.GetBackupRepoCleanup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// UpdateBackupRepoCleanupSettings Updates backup repo cleanup settings
// @Summary Updates backup repo cleanup settings
// @Description Updates backup repo cleanup settings
// @Tags    system_settings
// @Produce json
// @Id      updateBackupRepoCleanupSettings
// @Param   body body backuprepocleanupdto.UpdateBackupRepoCleanupReq true "request data"
// @Success 200 {object} backuprepocleanupdto.UpdateBackupRepoCleanupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/backup-repo-cleanup [put]
func (h *Handler) UpdateBackupRepoCleanupSettings(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeBackupRepoCleanup,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := backuprepocleanupdto.NewUpdateBackupRepoCleanupReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.BackupRepoCleanupUC.UpdateBackupRepoCleanup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ExecuteBackupRepoCleanup Executes the renewal
// @Summary Executes the renewal
// @Description Executes the renewal
// @Tags    system_settings
// @Produce json
// @Id      executeBackupRepoCleanup
// @Param   body body backuprepocleanupdto.ExecuteBackupRepoCleanupReq true "request data"
// @Success 200 {object} backuprepocleanupdto.ExecuteBackupRepoCleanupResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /system/settings/backup-repo-cleanup/exec [post]
func (h *Handler) ExecuteBackupRepoCleanup(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.GeneralResourceAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeExecute},
		Module:          base.ResourceModuleSystem,
		ResourceType:    base.ResourceTypeBackupRepoCleanup,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := backuprepocleanupdto.NewExecuteBackupRepoCleanupReq()
	req.Scope = entity.NewObjectScopeGlobal()
	if err = h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.BackupRepoCleanupUC.ExecuteBackupRepoCleanup(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
