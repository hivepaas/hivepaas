package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
)

// ListBasicAuth Lists basic auth settings
// @Summary Lists basic auth settings
// @Description Lists basic auth settings
// @Tags    settings
// @Produce json
// @Id      listSettingBasicAuth
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} basicauthdto.ListBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/basic-auth [get]
func (h *Handler) ListBasicAuth(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeBasicAuth, base.ObjectScopeGlobal)
}

// GetBasicAuth Gets basic auth setting details
// @Summary Gets basic auth setting details
// @Description Gets basic auth setting details
// @Tags    settings
// @Produce json
// @Id      getSettingBasicAuth
// @Param   itemID path string true "setting ID"
// @Success 200 {object} basicauthdto.GetBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/basic-auth/{itemID} [get]
func (h *Handler) GetBasicAuth(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeBasicAuth, base.ObjectScopeGlobal)
}

// CreateBasicAuth Creates a new basic auth setting
// @Summary Creates a new basic auth setting
// @Description Creates a new basic auth setting
// @Tags    settings
// @Produce json
// @Id      createSettingBasicAuth
// @Param   body body basicauthdto.CreateBasicAuthReq true "request data"
// @Success 201 {object} basicauthdto.CreateBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/basic-auth [post]
func (h *Handler) CreateBasicAuth(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeBasicAuth, base.ObjectScopeGlobal)
}

// UpdateBasicAuth Updates basic auth
// @Summary Updates basic auth
// @Description Updates basic auth
// @Tags    settings
// @Produce json
// @Id      updateSettingBasicAuth
// @Param   itemID path string true "setting ID"
// @Param   body body basicauthdto.UpdateBasicAuthReq true "request data"
// @Success 200 {object} basicauthdto.UpdateBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/basic-auth/{itemID} [put]
func (h *Handler) UpdateBasicAuth(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeBasicAuth, base.ObjectScopeGlobal)
}

// UpdateBasicAuthStatus Updates basic auth status
// @Summary Updates basic auth status
// @Description Updates basic auth status
// @Tags    settings
// @Produce json
// @Id      updateSettingBasicAuthStatus
// @Param   itemID path string true "setting ID"
// @Param   body body basicauthdto.UpdateBasicAuthStatusReq true "request data"
// @Success 200 {object} basicauthdto.UpdateBasicAuthStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/basic-auth/{itemID}/status [put]
func (h *Handler) UpdateBasicAuthStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeBasicAuth, base.ObjectScopeGlobal)
}

// DeleteBasicAuth Deletes basic auth setting
// @Summary Deletes basic auth setting
// @Description Deletes basic auth setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingBasicAuth
// @Param   itemID path string true "setting ID"
// @Success 200 {object} basicauthdto.DeleteBasicAuthResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/basic-auth/{itemID} [delete]
func (h *Handler) DeleteBasicAuth(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeBasicAuth, base.ObjectScopeGlobal)
}
