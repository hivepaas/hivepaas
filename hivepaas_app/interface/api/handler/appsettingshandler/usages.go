package appsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// AppSettingUsages answers what still references one of the app's settings -
// a secret or a config file a setting mount reads, a job a sequence runs:
// SettingUsages, under each settings group of the app. The handler is taken
// when a request comes, not when the route is registered.
//
// @Summary Lists what references an app's setting
// @Description Lists what still references one of the app's settings: the apps and settings that use it, such as
// @Description the setting mount that reads a secret. Deleting a setting they use fails with ERR_SETTING_IN_USE. kind
// @Description is the settings group, as in its other routes: secrets, config-files, setting-mounts, sched-jobs or
// @Description periodic-jobs.
// @Tags    App settings
// @Produce json
// @Id      getAppSettingUsages
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project env"
// @Param   appID path string true "app ID"
// @Param   kind path string true "the settings group, such as secrets"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} basesettinghandler.GetSettingUsagesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/apps/{appID}/{kind}/{itemID}/usages [get]
func (h *Handler) AppSettingUsages(resType base.ResourceType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h.SettingUsages(resType, base.ObjectScopeApp)(ctx)
	}
}
