package projectsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// ProjectSettingUsages answers what still references one of the project's
// settings: SettingUsages, under each settings group of the project. The
// handler is taken when a request comes, not when the route is registered.
//
// @Summary Lists what references a project's setting
// @Description Lists what still references one of the project's settings: the apps, projects and settings that
// @Description use it. Deleting a setting they use fails with ERR_SETTING_IN_USE. kind is the settings group, as in
// @Description its other routes, such as ssh-keys, secrets or config-files.
// @Tags    Project settings
// @Produce json
// @Id      getProjectSettingUsages
// @Param   projectID path string true "project ID"
// @Param   kind path string true "the settings group, such as ssh-keys"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} basesettinghandler.GetSettingUsagesResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 404 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{kind}/{itemID}/usages [get]
func (h *Handler) ProjectSettingUsages(resType base.ResourceType) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		h.SettingUsages(resType, base.ObjectScopeProject)(ctx)
	}
}
