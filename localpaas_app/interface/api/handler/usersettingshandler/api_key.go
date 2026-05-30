package usersettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/basesettinghandler"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/usersettings/apikeyuc/apikeydto"
)

// ListAPIKey Lists API key
// @Summary Lists API key
// @Description Lists API key
// @Tags    user_settings
// @Produce json
// @Id      listUserAPIKey
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} apikeydto.ListAPIKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/settings/api-keys [get]
func (h *Handler) ListAPIKey(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeAPIKey, base.ObjectScopeUser)
}

// GetAPIKey Gets API key details
// @Summary Gets API key details
// @Description Gets API key details
// @Tags    user_settings
// @Produce json
// @Id      getUserAPIKey
// @Param   itemID path string true "setting ID"
// @Success 200 {object} apikeydto.GetAPIKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/settings/api-keys/{itemID} [get]
func (h *Handler) GetAPIKey(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeAPIKey, base.ObjectScopeUser)
}

// CreateAPIKey Creates a new API key
// @Summary Creates a new API key
// @Description Creates a new API key
// @Tags    user_settings
// @Produce json
// @Id      createUserAPIKey
// @Param   body body apikeydto.CreateAPIKeyReq true "request data"
// @Success 201 {object} apikeydto.CreateAPIKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/settings/api-keys [post]
func (h *Handler) CreateAPIKey(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeAPIKey, base.ObjectScopeUser,
		basesettinghandler.CreateSettingPreRequestHandler(func(auth *basedto.Auth, req any) error {
			// Not allow to use API key to create API key (TODO: improve this behavior?)
			if auth.User.AuthClaims.IsAPIKey {
				return apperrors.New(apperrors.ErrForbidden).
					WithMsgLog("not allow to create API key by using API key session")
			}
			return nil
		}))
}

// UpdateAPIKeyStatus Updates API key status
// @Summary Updates API key status
// @Description Updates API key status
// @Tags    user_settings
// @Produce json
// @Id      updateUserAPIKeyStatus
// @Param   itemID path string true "setting ID"
// @Param   body body apikeydto.UpdateAPIKeyStatusReq true "request data"
// @Success 200 {object} apikeydto.UpdateAPIKeyStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/settings/api-keys/{itemID}/status [put]
func (h *Handler) UpdateAPIKeyStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeAPIKey, base.ObjectScopeUser,
		basesettinghandler.UpdateSettingPreRequestHandler(func(auth *basedto.Auth, req any) error {
			// Not allow to use API key to update API key (TODO: improve this behavior?)
			if auth.User.AuthClaims.IsAPIKey {
				return apperrors.New(apperrors.ErrForbidden).
					WithMsgLog("not allow to update API key by using API key session")
			}
			return nil
		}))
}

// DeleteAPIKey Deletes an API key
// @Summary Deletes an API key
// @Description Deletes an API key
// @Tags    user_settings
// @Produce json
// @Id      deleteUserAPIKey
// @Param   itemID path string true "setting ID"
// @Success 200 {object} apikeydto.DeleteAPIKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /users/current/settings/api-keys/{itemID} [delete]
func (h *Handler) DeleteAPIKey(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeAPIKey, base.ObjectScopeUser,
		basesettinghandler.DeleteSettingPreRequestHandler(func(auth *basedto.Auth, req any) error {
			// Not allow to use API key to delete API key (TODO: improve this behavior?)
			if auth.User.AuthClaims.IsAPIKey {
				return apperrors.New(apperrors.ErrForbidden).
					WithMsgLog("not allow to delete API key by using API key session")
			}
			return nil
		}))
}
