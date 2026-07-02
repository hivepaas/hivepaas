package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslprovideruc/sslproviderdto"
)

// ListSSLProvider Lists SSL providers
// @Summary Lists SSL providers
// @Description Lists SSL providers
// @Tags    project_settings
// @Produce json
// @Id      listProjectSSLProvider
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} sslproviderdto.ListSSLProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssl-providers [get]
func (h *Handler) ListSSLProvider(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSSLProvider, base.ObjectScopeProject)
}

// GetSSLProvider Gets SSL provider details
// @Summary Gets SSL provider details
// @Description Gets SSL provider details
// @Tags    project_settings
// @Produce json
// @Id      getProjectSSLProvider
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sslproviderdto.GetSSLProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssl-providers/{itemID} [get]
func (h *Handler) GetSSLProvider(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSSLProvider, base.ObjectScopeProject)
}

// CreateSSLProvider Creates a new SSL provider
// @Summary Creates a new SSL provider
// @Description Creates a new SSL provider
// @Tags    project_settings
// @Produce json
// @Id      createProjectSSLProvider
// @Param   projectID path string true "project ID"
// @Param   body body sslproviderdto.CreateSSLProviderReq true "request data"
// @Success 201 {object} sslproviderdto.CreateSSLProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssl-providers [post]
func (h *Handler) CreateSSLProvider(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSSLProvider, base.ObjectScopeProject)
}

// UpdateSSLProvider Updates an SSL provider
// @Summary Updates an SSL provider
// @Description Updates an SSL provider
// @Tags    project_settings
// @Produce json
// @Id      updateProjectSSLProvider
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body sslproviderdto.UpdateSSLProviderReq true "request data"
// @Success 200 {object} sslproviderdto.UpdateSSLProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssl-providers/{itemID} [put]
func (h *Handler) UpdateSSLProvider(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSSLProvider, base.ObjectScopeProject)
}

// UpdateSSLProviderStatus Updates SSL provider status
// @Summary Updates SSL provider status
// @Description Updates SSL provider status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectSSLProviderStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body sslproviderdto.UpdateSSLProviderStatusReq true "request data"
// @Success 200 {object} sslproviderdto.UpdateSSLProviderStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssl-providers/{itemID}/status [put]
func (h *Handler) UpdateSSLProviderStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSSLProvider, base.ObjectScopeProject)
}

// DeleteSSLProvider Deletes an SSL provider
// @Summary Deletes an SSL provider
// @Description Deletes an SSL provider
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectSSLProvider
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} sslproviderdto.DeleteSSLProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/ssl-providers/{itemID} [delete]
func (h *Handler) DeleteSSLProvider(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSSLProvider, base.ObjectScopeProject)
}
