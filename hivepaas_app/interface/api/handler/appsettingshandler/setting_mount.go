package appsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/settingmountuc/settingmountdto"
)

// ListSettingMount Lists app setting mounts
// @Summary Lists app setting mounts
// @Description Lists app setting mounts
// @Tags    app_settings
// @Produce json
// @Id      listAppSettingMount
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} settingmountdto.ListSettingMountResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts [get]
func (h *Handler) ListSettingMount(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSettingMount, base.ObjectScopeApp)
}

// GetSettingMount Get an app setting mount details
// @Summary Get an app setting mount details
// @Description Get an app setting mount details
// @Tags    app_settings
// @Produce json
// @Id      getAppSettingMount
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} settingmountdto.GetSettingMountResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts/{itemID} [get]
func (h *Handler) GetSettingMount(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeSettingMount, base.ObjectScopeApp)
}

// CreateSettingMount Creates an app setting mount
// @Summary Creates an app setting mount
// @Description Creates an app setting mount
// @Tags    app_settings
// @Produce json
// @Id      createAppSettingMount
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   body body settingmountdto.CreateSettingMountReq true "request data"
// @Success 201 {object} settingmountdto.CreateSettingMountResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts [post]
func (h *Handler) CreateSettingMount(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSettingMount, base.ObjectScopeApp)
}

// UpdateSettingMount Updates an app setting mount
// @Summary Updates an app setting mount
// @Description Updates an app setting mount
// @Tags    app_settings
// @Produce json
// @Id      updateAppSettingMount
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body settingmountdto.UpdateSettingMountReq true "request data"
// @Success 200 {object} settingmountdto.UpdateSettingMountResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts/{itemID} [put]
func (h *Handler) UpdateSettingMount(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSettingMount, base.ObjectScopeApp)
}

// UpdateSettingMountStatus Updates app setting mount status
// @Summary Updates app setting mount status
// @Description Updates app setting mount status
// @Tags    app_settings
// @Produce json
// @Id      updateAppSettingMountStatus
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Param   body body settingmountdto.UpdateSettingMountStatusReq true "request data"
// @Success 200 {object} settingmountdto.UpdateSettingMountStatusResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts/{itemID}/status [put]
func (h *Handler) UpdateSettingMountStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSettingMount, base.ObjectScopeApp)
}

// DeleteSettingMount Deletes an app setting mount
// @Summary Deletes an app setting mount
// @Description Deletes an app setting mount
// @Tags    app_settings
// @Produce json
// @Id      deleteAppSettingMount
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} settingmountdto.DeleteSettingMountResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts/{itemID} [delete]
func (h *Handler) DeleteSettingMount(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSettingMount, base.ObjectScopeApp)
}

// ListSettingMountSources Lists the settings an app can mount, and their parts
// @Summary Lists the settings an app can mount, and their parts
// @Description Lists the settings an app can mount, and their parts
// @Tags    app_settings
// @Produce json
// @Id      listAppSettingMountSources
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Success 200 {object} settingmountdto.ListSettingMountSourcesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/setting-mounts/sources [get]
func (h *Handler) ListSettingMountSources(ctx *gin.Context) {
	// Read on the app is what is asked; the sources do not depend on which app.
	auth, _, _, _, err := h.GetAuth(ctx, base.ActionTypeRead) //nolint:dogsled
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	resp, err := h.SettingMountUC.ListSettingMountSources(h.RequestCtx(ctx), auth,
		settingmountdto.NewListSettingMountSourcesReq())
	if err != nil {
		h.RenderError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}
