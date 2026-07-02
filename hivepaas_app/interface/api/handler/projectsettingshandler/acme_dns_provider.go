package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/acmednsprovideruc/acmednsproviderdto"
)

// ListAcmeDnsProvider Lists ACME DNS providers
// @Summary Lists ACME DNS providers
// @Description Lists ACME DNS providers
// @Tags    project_settings
// @Produce json
// @Id      listProjectAcmeDnsProvider
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} acmednsproviderdto.ListAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/acme-dns-providers [get]
func (h *Handler) ListAcmeDnsProvider(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeProject)
}

// GetAcmeDnsProvider Gets ACME DNS provider details
// @Summary Gets ACME DNS provider details
// @Description Gets ACME DNS provider details
// @Tags    project_settings
// @Produce json
// @Id      getProjectAcmeDnsProvider
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} acmednsproviderdto.GetAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/acme-dns-providers/{itemID} [get]
func (h *Handler) GetAcmeDnsProvider(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeProject)
}

// CreateAcmeDnsProvider Creates a new ACME DNS provider
// @Summary Creates a new ACME DNS provider
// @Description Creates a new ACME DNS provider
// @Tags    project_settings
// @Produce json
// @Id      createProjectAcmeDnsProvider
// @Param   projectID path string true "project ID"
// @Param   body body acmednsproviderdto.CreateAcmeDnsProviderReq true "request data"
// @Success 201 {object} acmednsproviderdto.CreateAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/acme-dns-providers [post]
func (h *Handler) CreateAcmeDnsProvider(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeProject)
}

// UpdateAcmeDnsProvider Updates an ACME DNS provider
// @Summary Updates an ACME DNS provider
// @Description Updates an ACME DNS provider
// @Tags    project_settings
// @Produce json
// @Id      updateProjectAcmeDnsProvider
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body acmednsproviderdto.UpdateAcmeDnsProviderReq true "request data"
// @Success 200 {object} acmednsproviderdto.UpdateAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/acme-dns-providers/{itemID} [put]
func (h *Handler) UpdateAcmeDnsProvider(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeProject)
}

// UpdateAcmeDnsProviderStatus Updates ACME DNS provider status
// @Summary Updates ACME DNS provider status
// @Description Updates ACME DNS provider status
// @Tags    project_settings
// @Produce json
// @Id      updateProjectAcmeDnsProviderStatus
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Param   body body acmednsproviderdto.UpdateAcmeDnsProviderStatusReq true "request data"
// @Success 200 {object} acmednsproviderdto.UpdateAcmeDnsProviderStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/acme-dns-providers/{itemID}/status [put]
func (h *Handler) UpdateAcmeDnsProviderStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeProject)
}

// DeleteAcmeDnsProvider Deletes an ACME DNS provider
// @Summary Deletes an ACME DNS provider
// @Description Deletes an ACME DNS provider
// @Tags    project_settings
// @Produce json
// @Id      deleteProjectAcmeDnsProvider
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} acmednsproviderdto.DeleteAcmeDnsProviderResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/acme-dns-providers/{itemID} [delete]
func (h *Handler) DeleteAcmeDnsProvider(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeAcmeDnsProvider, base.ObjectScopeProject)
}
