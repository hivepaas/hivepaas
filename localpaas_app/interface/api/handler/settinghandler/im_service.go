package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler/authhandler"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imserviceuc/imservicedto"
)

// ListIMService Lists IM services
// @Summary Lists IM services
// @Description Lists IM services
// @Tags    settings
// @Produce json
// @Id      listSettingIMService
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} imservicedto.ListIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/im-services [get]
func (h *Handler) ListIMService(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeGlobal)
}

// GetIMService Gets IM service details
// @Summary Gets IM service details
// @Description Gets IM service details
// @Tags    settings
// @Produce json
// @Id      getSettingIMService
// @Param   itemID path string true "setting ID"
// @Success 200 {object} imservicedto.GetIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/im-services/{itemID} [get]
func (h *Handler) GetIMService(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeGlobal)
}

// CreateIMService Creates a new IM service
// @Summary Creates a new IM service
// @Description Creates a new IM service
// @Tags    settings
// @Produce json
// @Id      createSettingIMService
// @Param   body body imservicedto.CreateIMServiceReq true "request data"
// @Success 201 {object} imservicedto.CreateIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/im-services [post]
func (h *Handler) CreateIMService(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeGlobal)
}

// UpdateIMService Updates IM service
// @Summary Updates IM service
// @Description Updates IM service
// @Tags    settings
// @Produce json
// @Id      updateSettingIMService
// @Param   itemID path string true "setting ID"
// @Param   body body imservicedto.UpdateIMServiceReq true "request data"
// @Success 200 {object} imservicedto.UpdateIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/im-services/{itemID} [put]
func (h *Handler) UpdateIMService(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeGlobal)
}

// UpdateIMServiceStatus Updates IMService status setting
// @Summary Updates IMService status setting
// @Description Updates IMService status setting
// @Tags    settings
// @Produce json
// @Id      updateSettingIMServiceStatus
// @Param   itemID path string true "setting ID"
// @Param   body body imservicedto.UpdateIMServiceStatusReq true "request data"
// @Success 200 {object} imservicedto.UpdateIMServiceStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/im-services/{itemID}/status [put]
func (h *Handler) UpdateIMServiceStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeIMService, base.ObjectScopeGlobal)
}

// DeleteIMService Deletes IM service
// @Summary Deletes IM service
// @Description Deletes IM service
// @Tags    settings
// @Produce json
// @Id      deleteSettingIMService
// @Param   itemID path string true "setting ID"
// @Success 200 {object} imservicedto.DeleteIMServiceResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/im-services/{itemID} [delete]
func (h *Handler) DeleteIMService(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeIMService, base.ObjectScopeGlobal)
}

// TestSendInstantMsg Tests sending a msg
// @Summary Tests sending a msg
// @Description Tests sending a msg
// @Tags    settings
// @Produce json
// @Id      testSendIMServiceMsg
// @Param   body body imservicedto.TestSendInstantMsgReq true "request data"
// @Success 200 {object} imservicedto.TestSendInstantMsgResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/im-services/test-send-msg [post]
func (h *Handler) TestSendInstantMsg(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := imservicedto.NewTestSendInstantMsgReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.IMServiceUC.TestSendInstantMsg(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
