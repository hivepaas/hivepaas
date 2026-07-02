package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler/authhandler"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
)

// ListAcmeDnsProvider Lists ACME DNS providers
// @Summary Lists ACME DNS providers
// @Description Lists ACME DNS providers
// @Tags    settings
// @Produce json
// @Id      listSettingAcmeDnsProvider
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} acmednsproviderdto.ListAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/acme-dns-providers [get]
func (h *Handler) ListAcmeDnsProvider(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeGlobal)
}

// GetAcmeDnsProvider Gets ACME DNS provider details
// @Summary Gets ACME DNS provider details
// @Description Gets ACME DNS provider details
// @Tags    settings
// @Produce json
// @Id      getSettingAcmeDnsProvider
// @Param   itemID path string true "setting ID"
// @Success 200 {object} acmednsproviderdto.GetAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/acme-dns-providers/{itemID} [get]
func (h *Handler) GetAcmeDnsProvider(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeGlobal)
}

// CreateAcmeDnsProvider Creates a new ACME DNS provider
// @Summary Creates a new ACME DNS provider
// @Description Creates a new ACME DNS provider
// @Tags    settings
// @Produce json
// @Id      createSettingAcmeDnsProvider
// @Param   body body acmednsproviderdto.CreateAcmeDnsProviderReq true "request data"
// @Success 201 {object} acmednsproviderdto.CreateAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/acme-dns-providers [post]
func (h *Handler) CreateAcmeDnsProvider(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeGlobal)
}

// UpdateAcmeDnsProvider Updates ACME DNS provider
// @Summary Updates ACME DNS provider
// @Description Updates ACME DNS provider
// @Tags    settings
// @Produce json
// @Id      updateSettingAcmeDnsProvider
// @Param   itemID path string true "setting ID"
// @Param   body body acmednsproviderdto.UpdateAcmeDnsProviderReq true "request data"
// @Success 200 {object} acmednsproviderdto.UpdateAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/acme-dns-providers/{itemID} [put]
func (h *Handler) UpdateAcmeDnsProvider(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeGlobal)
}

// UpdateAcmeDnsProviderStatus Updates ACME DNS provider status
// @Summary Updates ACME DNS provider status
// @Description Updates ACME DNS provider status
// @Tags    settings
// @Produce json
// @Id      updateSettingAcmeDnsProviderStatus
// @Param   itemID path string true "setting ID"
// @Param   body body acmednsproviderdto.UpdateAcmeDnsProviderStatusReq true "request data"
// @Success 200 {object} acmednsproviderdto.UpdateAcmeDnsProviderStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/acme-dns-providers/{itemID}/status [put]
func (h *Handler) UpdateAcmeDnsProviderStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeGlobal)
}

// DeleteAcmeDnsProvider Deletes ACME DNS provider
// @Summary Deletes ACME DNS provider
// @Description Deletes ACME DNS provider
// @Tags    settings
// @Produce json
// @Id      deleteSettingAcmeDnsProvider
// @Param   itemID path string true "setting ID"
// @Success 200 {object} acmednsproviderdto.DeleteAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/acme-dns-providers/{itemID} [delete]
func (h *Handler) DeleteAcmeDnsProvider(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeGlobal)
}

// TestAcmeDnsProviderAccess Tests provider access
// @Summary Tests provider access
// @Description Tests provider access
// @Tags    settings
// @Produce json
// @Id      testAcmeDnsProviderAccess
// @Param   body body acmednsproviderdto.TestProviderAccessReq true "request data"
// @Success 200 {object} acmednsproviderdto.TestProviderAccessResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/acme-dns-providers/test-access [post]
func (h *Handler) TestAcmeDnsProviderAccess(ctx *gin.Context) {
	auth, err := h.AuthHandler.GetCurrentAuth(ctx, authhandler.NoAccessCheck)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := acmednsproviderdto.NewTestProviderAccessReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.AcmeDnsProviderUC.TestProviderAccess(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
