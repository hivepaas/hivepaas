package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// ProjectEnvSettingUsages answers what still references one of the env's
// settings: SettingUsages, under each settings group of the env. The handler
// is taken when a request comes, not when the route is registered.
//
// @Summary Lists what references an env's setting
// @Description Lists what still references one of the env's settings: the apps, projects and settings that use
// @Description it. Deleting a setting they use fails with ERR_SETTING_IN_USE. kind is the settings group, as in its
// @Description other routes, such as ssh-keys, secrets or config-files.
// @Tags    Project env settings
// @Produce json
// @Id      getProjectEnvSettingUsages
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   kind path string true "the settings group, such as ssh-keys"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} basesettinghandler.GetSettingUsagesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/{kind}/{itemID}/usages [get]
func (h *Handler) ProjectEnvSettingUsages(resType base.ResourceType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h.SettingUsages(resType, base.ObjectScopeProjectEnv)(ctx)
	}
}
