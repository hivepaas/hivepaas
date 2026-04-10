package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc/sslcertdto"
)

// ListSSLCert Lists SSL certs
// @Summary Lists SSL certs
// @Description Lists SSL certs
// @Tags    settings
// @Produce json
// @Id      listSettingSSLCert
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} sslcertdto.ListSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssl-certs [get]
func (h *Handler) ListSSLCert(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSSLCert, base.SettingScopeGlobal)
}

// GetSSLCert Gets SSL cert details
// @Summary Gets SSL cert details
// @Description Gets SSL cert details
// @Tags    settings
// @Produce json
// @Id      getSettingSSLCert
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sslcertdto.GetSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssl-certs/{itemID} [get]
func (h *Handler) GetSSLCert(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSSLCert, base.SettingScopeGlobal)
}

// CreateSSLCert Creates a new SSL cert
// @Summary Creates a new SSL cert
// @Description Creates a new SSL cert
// @Tags    settings
// @Produce json
// @Id      createSettingSSLCert
// @Param   body body sslcertdto.CreateSSLCertReq true "request data"
// @Success 201 {object} sslcertdto.CreateSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssl-certs [post]
func (h *Handler) CreateSSLCert(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSSLCert, base.SettingScopeGlobal)
}

// UpdateSSLCert Updates SSL cert
// @Summary Updates SSL cert
// @Description Updates SSL cert
// @Tags    settings
// @Produce json
// @Id      updateSettingSSLCert
// @Param   itemID path string true "setting ID"
// @Param   body body sslcertdto.UpdateSSLCertReq true "request data"
// @Success 200 {object} sslcertdto.UpdateSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssl-certs/{itemID} [put]
func (h *Handler) UpdateSSLCert(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSSLCert, base.SettingScopeGlobal)
}

// UpdateSSLCertStatus Updates SSL cert status
// @Summary Updates SSL cert status
// @Description Updates SSL cert status
// @Tags    settings
// @Produce json
// @Id      updateSettingSSLCertStatus
// @Param   itemID path string true "setting ID"
// @Param   body body sslcertdto.UpdateSSLCertStatusReq true "request data"
// @Success 200 {object} sslcertdto.UpdateSSLCertStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssl-certs/{itemID}/status [put]
func (h *Handler) UpdateSSLCertStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSSLCert, base.SettingScopeGlobal)
}

// DeleteSSLCert Deletes SSL cert
// @Summary Deletes SSL cert
// @Description Deletes SSL cert
// @Tags    settings
// @Produce json
// @Id      deleteSettingSSLCert
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sslcertdto.DeleteSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/ssl-certs/{itemID} [delete]
func (h *Handler) DeleteSSLCert(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSSLCert, base.SettingScopeGlobal)
}
