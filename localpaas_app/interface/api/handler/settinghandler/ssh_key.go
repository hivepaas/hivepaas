package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sshkeyuc/sshkeydto"
)

// ListSSHKey Lists ssh-key settings
// @Summary Lists ssh-key settings
// @Description Lists ssh-key settings
// @Tags    settings
// @Produce json
// @Id      listSettingSSHKey
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} sshkeydto.ListSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssh-keys [get]
func (h *Handler) ListSSHKey(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSSHKey, base.SettingScopeGlobal)
}

// GetSSHKey Gets ssh-key setting details
// @Summary Gets ssh-key setting details
// @Description Gets ssh-key setting details
// @Tags    settings
// @Produce json
// @Id      getSettingSSHKey
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sshkeydto.GetSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssh-keys/{itemID} [get]
func (h *Handler) GetSSHKey(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSSHKey, base.SettingScopeGlobal)
}

// CreateSSHKey Creates a new ssh-key setting
// @Summary Creates a new ssh-key setting
// @Description Creates a new ssh-key setting
// @Tags    settings
// @Produce json
// @Id      createSettingSSHKey
// @Param   body body sshkeydto.CreateSSHKeyReq true "request data"
// @Success 201 {object} sshkeydto.CreateSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssh-keys [post]
func (h *Handler) CreateSSHKey(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSSHKey, base.SettingScopeGlobal)
}

// UpdateSSHKey Updates ssh-key
// @Summary Updates ssh-key
// @Description Updates ssh-key
// @Tags    settings
// @Produce json
// @Id      updateSettingSSHKey
// @Param   itemID path string true "setting ID"
// @Param   body body sshkeydto.UpdateSSHKeyReq true "request data"
// @Success 200 {object} sshkeydto.UpdateSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssh-keys/{itemID} [put]
func (h *Handler) UpdateSSHKey(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSSHKey, base.SettingScopeGlobal)
}

// UpdateSSHKeyStatus Updates ssh-key status
// @Summary Updates ssh-key status
// @Description Updates ssh-key status
// @Tags    settings
// @Produce json
// @Id      updateSettingSSHKeyStatus
// @Param   itemID path string true "setting ID"
// @Param   body body sshkeydto.UpdateSSHKeyStatusReq true "request data"
// @Success 200 {object} sshkeydto.UpdateSSHKeyStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssh-keys/{itemID}/status [put]
func (h *Handler) UpdateSSHKeyStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSSHKey, base.SettingScopeGlobal)
}

// DeleteSSHKey Deletes sshkey setting
// @Summary Deletes sshkey setting
// @Description Deletes sshkey setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingSSHKey
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sshkeydto.DeleteSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssh-keys/{itemID} [delete]
func (h *Handler) DeleteSSHKey(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSSHKey, base.SettingScopeGlobal)
}

// GenerateSSHKey Generates an SSH key
// @Summary Generates an SSH key
// @Description Generates an SSH key
// @Tags    settings
// @Produce json
// @Id      generateSSHKey
// @Param   body body sshkeydto.GenerateSSHKeyReq true "request data"
// @Success 200 {object} sshkeydto.GenerateSSHKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssh-keys/generate [post]
func (h *Handler) GenerateSSHKey(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := sshkeydto.NewGenerateSSHKeyReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.SSHKeyUC.GenerateSSHKey(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
