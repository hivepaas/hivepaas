package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
)

// ListRegistryAuth Lists registry auth settings
// @Summary Lists registry auth settings
// @Description Lists registry auth settings
// @Tags    settings
// @Produce json
// @Id      listSettingRegistryAuth
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} registryauthdto.ListRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/registry-auth [get]
func (h *Handler) ListRegistryAuth(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeGlobal)
}

// GetRegistryAuth Gets registry auth setting details
// @Summary Gets registry auth setting details
// @Description Gets registry auth setting details
// @Tags    settings
// @Produce json
// @Id      getSettingRegistryAuth
// @Param   itemID path string true "setting ID"
// @Success 200 {object} registryauthdto.GetRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/registry-auth/{itemID} [get]
func (h *Handler) GetRegistryAuth(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeGlobal)
}

// CreateRegistryAuth Creates a new registry auth setting
// @Summary Creates a new registry auth setting
// @Description Creates a new registry auth setting
// @Tags    settings
// @Produce json
// @Id      createSettingRegistryAuth
// @Param   body body registryauthdto.CreateRegistryAuthReq true "request data"
// @Success 201 {object} registryauthdto.CreateRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/registry-auth [post]
func (h *Handler) CreateRegistryAuth(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeGlobal)
}

// UpdateRegistryAuth Updates registry auth
// @Summary Updates registry auth
// @Description Updates registry auth
// @Tags    settings
// @Produce json
// @Id      updateSettingRegistryAuth
// @Param   itemID path string true "setting ID"
// @Param   body body registryauthdto.UpdateRegistryAuthReq true "request data"
// @Success 200 {object} registryauthdto.UpdateRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/registry-auth/{itemID} [put]
func (h *Handler) UpdateRegistryAuth(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeGlobal)
}

// UpdateRegistryAuthStatus Updates registry auth status
// @Summary Updates registry auth status
// @Description Updates registry auth status
// @Tags    settings
// @Produce json
// @Id      updateSettingRegistryAuthStatus
// @Param   itemID path string true "setting ID"
// @Param   body body registryauthdto.UpdateRegistryAuthStatusReq true "request data"
// @Success 200 {object} registryauthdto.UpdateRegistryAuthStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/registry-auth/{itemID}/status [put]
func (h *Handler) UpdateRegistryAuthStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeGlobal)
}

// DeleteRegistryAuth Deletes registry auth setting
// @Summary Deletes registry auth setting
// @Description Deletes registry auth setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingRegistryAuth
// @Param   itemID path string true "setting ID"
// @Success 200 {object} registryauthdto.DeleteRegistryAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/registry-auth/{itemID} [delete]
func (h *Handler) DeleteRegistryAuth(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeRegistryAuth, base.ObjectScopeGlobal)
}

// TestRegistryAuthConn Tests registry auth connection
// @Summary Tests registry auth connection
// @Description Tests registry auth connection
// @Tags    settings
// @Produce json
// @Id      testRegistryAuthConn
// @Param   body body registryauthdto.TestRegistryAuthConnReq true "request data"
// @Success 200 {object} registryauthdto.TestRegistryAuthConnResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/registry-auth/test-conn [post]
func (h *Handler) TestRegistryAuthConn(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := registryauthdto.NewTestRegistryAuthConnReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.RegistryAuthUC.TestRegistryAuthConn(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
