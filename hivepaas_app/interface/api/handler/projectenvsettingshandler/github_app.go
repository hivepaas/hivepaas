package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/githubappuc/githubappdto"
)

// ListGithubApp Lists github-app settings
// @Summary Lists github-app settings
// @Description Lists github-app settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvGithubApp
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} githubappdto.ListGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/github-apps [get]
func (h *Handler) ListGithubApp(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeGithubApp, base.ObjectScopeProjectEnv)
}

// GetGithubApp Gets github-app setting details
// @Summary Gets github-app setting details
// @Description Gets github-app setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvGithubApp
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} githubappdto.GetGithubAppResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/github-apps/{itemID} [get]
func (h *Handler) GetGithubApp(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeGithubApp, base.ObjectScopeProjectEnv)
}
