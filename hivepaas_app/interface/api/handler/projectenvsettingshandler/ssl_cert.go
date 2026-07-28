package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

// ListSSLCert Lists SSL certs
// @Summary Lists SSL certs
// @Description Lists SSL certs
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvSSLCert
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} sslcertdto.ListSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssl-certs [get]
func (h *Handler) ListSSLCert(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSSLCert, base.ObjectScopeProjectEnv)
}

// GetSSLCert Gets SSL cert details
// @Summary Gets SSL cert details
// @Description Gets SSL cert details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvSSLCert
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sslcertdto.GetSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssl-certs/{itemID} [get]
func (h *Handler) GetSSLCert(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSSLCert, base.ObjectScopeProjectEnv)
}

// CreateSSLCert Creates a new SSL cert
// @Summary Creates a new SSL cert
// @Description Creates a new SSL cert
// @Tags    project_env_settings
// @Produce json
// @Id      createProjectEnvSSLCert
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   body body sslcertdto.CreateSSLCertReq true "request data"
// @Success 201 {object} sslcertdto.CreateSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssl-certs [post]
func (h *Handler) CreateSSLCert(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSSLCert, base.ObjectScopeProjectEnv)
}

// UpdateSSLCert Updates an SSL cert
// @Summary Updates an SSL cert
// @Description Updates an SSL cert
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvSSLCert
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body sslcertdto.UpdateSSLCertReq true "request data"
// @Success 200 {object} sslcertdto.UpdateSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssl-certs/{itemID} [put]
func (h *Handler) UpdateSSLCert(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSSLCert, base.ObjectScopeProjectEnv)
}

// UpdateSSLCertStatus Updates SSL cert status
// @Summary Updates SSL cert status
// @Description Updates SSL cert status
// @Tags    project_env_settings
// @Produce json
// @Id      updateProjectEnvSSLCertStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body sslcertdto.UpdateSSLCertStatusReq true "request data"
// @Success 200 {object} sslcertdto.UpdateSSLCertStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssl-certs/{itemID}/status [put]
func (h *Handler) UpdateSSLCertStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSSLCert, base.ObjectScopeProjectEnv)
}

// DeleteSSLCert Deletes an SSL cert
// @Summary Deletes an SSL cert
// @Description Deletes an SSL cert
// @Tags    project_env_settings
// @Produce json
// @Id      deleteProjectEnvSSLCert
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sslcertdto.DeleteSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssl-certs/{itemID} [delete]
func (h *Handler) DeleteSSLCert(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSSLCert, base.ObjectScopeProjectEnv)
}

// RenewSSLCert Renews SSL cert
// @Summary Renews SSL cert
// @Description Renews SSL cert
// @Tags    project_env_settings
// @Produce json
// @Id      renewProjectEnvSSLCert
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Param   body body sslcertdto.RenewSSLCertReq true "request data"
// @Success 200 {object} sslcertdto.RenewSSLCertResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/ssl-certs/{itemID}/renew [post]
func (h *Handler) RenewSSLCert(ctx *gin.Context) {
	h.SSLCertRenew(ctx, base.ObjectScopeProjectEnv)
}
