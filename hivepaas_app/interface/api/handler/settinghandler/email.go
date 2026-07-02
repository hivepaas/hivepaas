package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc/emaildto"
)

// ListEmail Lists e-mail settings
// @Summary Lists e-mail settings
// @Description Lists e-mail settings
// @Tags    settings
// @Produce json
// @Id      listSettingEmail
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} emaildto.ListEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/emails [get]
func (h *Handler) ListEmail(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeGlobal)
}

// GetEmail Gets e-mail setting details
// @Summary Gets e-mail setting details
// @Description Gets e-mail setting details
// @Tags    settings
// @Produce json
// @Id      getSettingEmail
// @Param   itemID path string true "setting ID"
// @Success 200 {object} emaildto.GetEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/emails/{itemID} [get]
func (h *Handler) GetEmail(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeGlobal)
}

// CreateEmail Creates a new e-mail setting
// @Summary Creates a new e-mail setting
// @Description Creates a new e-mail setting
// @Tags    settings
// @Produce json
// @Id      createSettingEmail
// @Param   body body emaildto.CreateEmailReq true "request data"
// @Success 201 {object} emaildto.CreateEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/emails [post]
func (h *Handler) CreateEmail(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeGlobal)
}

// UpdateEmail Updates e-mail setting
// @Summary Updates e-mail setting
// @Description Updates e-mail setting
// @Tags    settings
// @Produce json
// @Id      updateSettingEmail
// @Param   itemID path string true "setting ID"
// @Param   body body emaildto.UpdateEmailReq true "request data"
// @Success 200 {object} emaildto.UpdateEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/emails/{itemID} [put]
func (h *Handler) UpdateEmail(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeGlobal)
}

// UpdateEmailStatus Updates email setting status
// @Summary Updates email setting status
// @Description Updates email setting status
// @Tags    settings
// @Produce json
// @Id      updateSettingEmailStatus
// @Param   itemID path string true "setting ID"
// @Param   body body emaildto.UpdateEmailStatusReq true "request data"
// @Success 200 {object} emaildto.UpdateEmailStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/emails/{itemID}/status [put]
func (h *Handler) UpdateEmailStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeEmail, base.ObjectScopeGlobal)
}

// DeleteEmail Deletes e-mail setting
// @Summary Deletes e-mail setting
// @Description Deletes e-mail setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingEmail
// @Param   itemID path string true "setting ID"
// @Success 200 {object} emaildto.DeleteEmailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/emails/{itemID} [delete]
func (h *Handler) DeleteEmail(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeEmail, base.ObjectScopeGlobal)
}

// TestSendMail Tests sending an email
// @Summary Tests sending an email
// @Description Tests sending an email
// @Tags    settings
// @Produce json
// @Id      testSendMail
// @Param   body body emaildto.TestSendMailReq true "request data"
// @Success 200 {object} emaildto.TestSendMailResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/emails/test-send-mail [post]
func (h *Handler) TestSendMail(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := emaildto.NewTestSendMailReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.EmailUC.TestSendMail(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
