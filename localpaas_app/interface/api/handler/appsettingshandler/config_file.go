package appsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/configfileuc/configfiledto"
)

// ListConfigFile Lists app config files
// @Summary Lists app config files
// @Description Lists app config files
// @Tags    app_settings
// @Produce json
// @Id      listAppConfigFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Success 200 {object} configfiledto.ListConfigFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files [get]
func (h *Handler) ListConfigFile(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp)
}

// GetConfigFile Get an app config file details
// @Summary Get an app config file details
// @Description Get an app config file details
// @Tags    app_settings
// @Produce json
// @Id      getAppConfigFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} configfiledto.GetConfigFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files/{itemID} [get]
func (h *Handler) GetConfigFile(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp)
}

// GetConfigFileDownloadToken Gets config file download token
// @Summary Gets config file download token
// @Description Gets config file download token
// @Tags    app_settings
// @Produce json
// @Id      getAppConfigFileDownloadToken
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} configfiledto.GetDownloadTokenResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files/{itemID}/download-token [get]
func (h *Handler) GetConfigFileDownloadToken(ctx *gin.Context) {
	h.GetDownloadToken(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp, "", 0)
}

// DownloadConfigFile Download a config file
// @Summary Download a config file
// @Description Download a config file
// @Tags    app_settings
// @Produce json
// @Id      downloadAppConfigFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} configfiledto.DownloadConfigFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files/{itemID}/download [get]
func (h *Handler) DownloadConfigFile(ctx *gin.Context) {
	h.Download(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp, "")
}

// CreateConfigFile Creates an app config file
// @Summary Creates an app config file
// @Description Creates an app config file
// @Tags    app_settings
// @Produce json
// @Id      createAppConfigFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   body body configfiledto.CreateConfigFileReq true "request data"
// @Success 201 {object} configfiledto.CreateConfigFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files [post]
func (h *Handler) CreateConfigFile(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp)
}

// UpdateConfigFile Updates an app config file
// @Summary Updates an app config file
// @Description Updates an app config file
// @Tags    app_settings
// @Produce json
// @Id      updateAppConfigFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body configfiledto.UpdateConfigFileReq true "request data"
// @Success 200 {object} configfiledto.UpdateConfigFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files/{itemID} [put]
func (h *Handler) UpdateConfigFile(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp)
}

// UpdateConfigFileStatus Updates app config file status
// @Summary Updates app config file status
// @Description Updates app config file status
// @Tags    app_settings
// @Produce json
// @Id      updateAppConfigFileStatus
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body configfiledto.UpdateConfigFileStatusReq true "request data"
// @Success 200 {object} configfiledto.UpdateConfigFileStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files/{itemID}/status [put]
func (h *Handler) UpdateConfigFileStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp)
}

// DeleteConfigFile Deletes an app config file
// @Summary Deletes an app config file
// @Description Deletes an app config file
// @Tags    app_settings
// @Produce json
// @Id      deleteAppConfigFile
// @Param   projectID path string true "project ID"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} configfiledto.DeleteConfigFileResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/apps/{appID}/config-files/{itemID} [delete]
func (h *Handler) DeleteConfigFile(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeConfigFile, base.ObjectScopeApp)
}
