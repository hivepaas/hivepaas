package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

// GetImageBuildSettings Gets image build settings
// @Summary Gets image build settings
// @Description Gets image build settings
// @Tags    settings
// @Produce json
// @Id      getSettingImageBuildSettings
// @Success 200 {object} imagebuildsettingsdto.GetImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings [get]
func (h *Handler) GetImageBuildSettings(ctx *gin.Context) {
	h.GetUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}

// UpdateImageBuildSettings Updates image build settings
// @Summary Updates image build settings
// @Description Updates image build settings
// @Tags    settings
// @Produce json
// @Id      updateSettingImageBuildSettings
// @Param   body body imagebuildsettingsdto.UpdateImageBuildSettingsReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings [put]
func (h *Handler) UpdateImageBuildSettings(ctx *gin.Context) {
	h.UpdateUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}

// UpdateImageBuildSettingsStatus Updates image build settings status
// @Summary Updates image build settings status
// @Description Updates image build settings status
// @Tags    settings
// @Produce json
// @Id      updateSettingImageBuildSettingsStatus
// @Param   body body imagebuildsettingsdto.UpdateImageBuildSettingsStatusReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.UpdateImageBuildSettingsStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings/status [put]
func (h *Handler) UpdateImageBuildSettingsStatus(ctx *gin.Context) {
	h.UpdateUniqueSettingStatus(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}

// DeleteImageBuildSettings Deletes image build settings
// @Summary Deletes image build settings
// @Description Deletes image build settings
// @Tags    settings
// @Produce json
// @Id      deleteSettingImageBuildSettings
// @Success 200 {object} imagebuildsettingsdto.DeleteImageBuildSettingsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings [delete]
func (h *Handler) DeleteImageBuildSettings(ctx *gin.Context) {
	h.DeleteUniqueSetting(ctx, base.ResourceTypeImageBuildSettings, base.ObjectScopeGlobal)
}

// GetRepoCacheInfo Gets repo cache info
// @Summary Gets repo cache info
// @Description Gets repo cache info
// @Tags    settings
// @Produce json
// @Id      getSettingRepoCacheInfo
// @Success 200 {object} imagebuildsettingsdto.GetRepoCacheInfoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings/repo-cache [get]
func (h *Handler) GetRepoCacheInfo(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSettings,
		Action:         base.ActionTypeRead,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagebuildsettingsdto.NewGetRepoCacheInfoReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageBuildUC.GetRepoCacheInfo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ClearRepoCache Clears repo cache
// @Summary Clears repo cache
// @Description Clears repo cache
// @Tags    settings
// @Produce json
// @Id      clearSettingRepoCache
// @Param   body body imagebuildsettingsdto.ClearRepoCacheReq true "request data"
// @Success 200 {object} imagebuildsettingsdto.ClearRepoCacheResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/image-build-settings/repo-cache/clear [post]
func (h *Handler) ClearRepoCache(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, &permission.AccessCheck{
		ResourceModule: base.ResourceModuleSettings,
		Action:         base.ActionTypeExecute,
	})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imagebuildsettingsdto.NewClearRepoCacheReq()
	req.Scope = base.NewObjectScopeGlobal()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.ImageBuildUC.ClearRepoCache(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
