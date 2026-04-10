package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/oauthuc/oauthdto"
)

// ListOAuth Lists oauth settings
// @Summary Lists oauth settings
// @Description Lists oauth settings
// @Tags    settings
// @Produce json
// @Id      listSettingOAuth
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} oauthdto.ListOAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/oauth [get]
func (h *Handler) ListOAuth(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeOAuth, base.SettingScopeGlobal)
}

// GetOAuth Gets oauth setting details
// @Summary Gets oauth setting details
// @Description Gets oauth setting details
// @Tags    settings
// @Produce json
// @Id      getSettingOAuth
// @Param   itemID path string true "setting ID"
// @Success 200 {object} oauthdto.GetOAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/oauth/{itemID} [get]
func (h *Handler) GetOAuth(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeOAuth, base.SettingScopeGlobal)
}

// CreateOAuth Creates a new oauth setting
// @Summary Creates a new oauth setting
// @Description Creates a new oauth setting
// @Tags    settings
// @Produce json
// @Id      createSettingOAuth
// @Param   body body oauthdto.CreateOAuthReq true "request data"
// @Success 201 {object} oauthdto.CreateOAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/oauth [post]
func (h *Handler) CreateOAuth(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeOAuth, base.SettingScopeGlobal)
}

// UpdateOAuth Updates oauth
// @Summary Updates oauth
// @Description Updates oauth
// @Tags    settings
// @Produce json
// @Id      updateSettingOAuth
// @Param   itemID path string true "setting ID"
// @Param   body body oauthdto.UpdateOAuthReq true "request data"
// @Success 200 {object} oauthdto.UpdateOAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/oauth/{itemID} [put]
func (h *Handler) UpdateOAuth(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeOAuth, base.SettingScopeGlobal)
}

// UpdateOAuthStatus Updates oauth status
// @Summary Updates oauth status
// @Description Updates oauth status
// @Tags    settings
// @Produce json
// @Id      updateSettingOAuthStatus
// @Param   itemID path string true "setting ID"
// @Param   body body oauthdto.UpdateOAuthStatusReq true "request data"
// @Success 200 {object} oauthdto.UpdateOAuthStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/oauth/{itemID}/status [put]
func (h *Handler) UpdateOAuthStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeOAuth, base.SettingScopeGlobal)
}

// DeleteOAuth Deletes oauth setting
// @Summary Deletes oauth setting
// @Description Deletes oauth setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingOAuth
// @Param   itemID path string true "setting ID"
// @Success 200 {object} oauthdto.DeleteOAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/oauth/{itemID} [delete]
func (h *Handler) DeleteOAuth(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeOAuth, base.SettingScopeGlobal)
}
