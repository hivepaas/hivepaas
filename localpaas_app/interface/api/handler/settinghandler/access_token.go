package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/accesstokenuc/accesstokendto"
)

// ListAccessToken Lists access-token settings
// @Summary Lists access-token settings
// @Description Lists access-token settings
// @Tags    settings
// @Produce json
// @Id      listSettingAccessToken
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} accesstokendto.ListAccessTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/access-tokens [get]
func (h *Handler) ListAccessToken(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeAccessToken, base.ObjectScopeGlobal)
}

// GetAccessToken Gets access-token setting details
// @Summary Gets access-token setting details
// @Description Gets access-token setting details
// @Tags    settings
// @Produce json
// @Id      getSettingAccessToken
// @Param   itemID path string true "setting ID"
// @Success 200 {object} accesstokendto.GetAccessTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/access-tokens/{itemID} [get]
func (h *Handler) GetAccessToken(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeAccessToken, base.ObjectScopeGlobal)
}

// CreateAccessToken Creates a new access-token setting
// @Summary Creates a new access-token setting
// @Description Creates a new access-token setting
// @Tags    settings
// @Produce json
// @Id      createSettingAccessToken
// @Param   body body accesstokendto.CreateAccessTokenReq true "request data"
// @Success 201 {object} accesstokendto.CreateAccessTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/access-tokens [post]
func (h *Handler) CreateAccessToken(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeAccessToken, base.ObjectScopeGlobal)
}

// UpdateAccessToken Updates access-token
// @Summary Updates access-token
// @Description Updates access-token
// @Tags    settings
// @Produce json
// @Id      updateSettingAccessToken
// @Param   itemID path string true "setting ID"
// @Param   body body accesstokendto.UpdateAccessTokenReq true "request data"
// @Success 200 {object} accesstokendto.UpdateAccessTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/access-tokens/{itemID} [put]
func (h *Handler) UpdateAccessToken(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeAccessToken, base.ObjectScopeGlobal)
}

// UpdateAccessTokenStatus Updates access-token status
// @Summary Updates access-token status
// @Description Updates access-token status
// @Tags    settings
// @Produce json
// @Id      updateSettingAccessTokenStatus
// @Param   itemID path string true "setting ID"
// @Param   body body accesstokendto.UpdateAccessTokenStatusReq true "request data"
// @Success 200 {object} accesstokendto.UpdateAccessTokenStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/access-tokens/{itemID}/status [put]
func (h *Handler) UpdateAccessTokenStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeAccessToken, base.ObjectScopeGlobal)
}

// DeleteAccessToken Deletes access-token setting
// @Summary Deletes access-token setting
// @Description Deletes access-token setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingAccessToken
// @Param   itemID path string true "setting ID"
// @Success 200 {object} accesstokendto.DeleteAccessTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/access-tokens/{itemID} [delete]
func (h *Handler) DeleteAccessToken(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeAccessToken, base.ObjectScopeGlobal)
}

// TestAccessTokenConn Test access-token connection
// @Summary Test access-token connection
// @Description Test access-token connection
// @Tags    settings
// @Produce json
// @Id      testAccessTokenConn
// @Param   body body accesstokendto.TestAccessTokenConnReq true "request data"
// @Success 200 {object} accesstokendto.TestAccessTokenConnResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/access-tokens/test-conn [post]
func (h *Handler) TestAccessTokenConn(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := accesstokendto.NewTestAccessTokenConnReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.AccessTokenUC.TestAccessTokenConn(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
