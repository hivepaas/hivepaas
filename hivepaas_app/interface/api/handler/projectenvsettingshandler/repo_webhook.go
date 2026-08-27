package projectenvsettingshandler

import (
	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

// ListRepoWebhook Lists webhook settings
// @Summary Lists webhook settings
// @Description Lists webhook settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvRepoWebhook
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} repowebhookdto.ListRepoWebhookResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/repo-webhooks [get]
func (h *Handler) ListRepoWebhook(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeProjectEnv)
}

// GetRepoWebhook Gets webhook setting details
// @Summary Gets webhook setting details
// @Description Gets webhook setting details
// @Tags    project_env_settings
// @Produce json
// @Id      getProjectEnvRepoWebhook
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "setting ID"
// @Success 200 {object} repowebhookdto.GetRepoWebhookResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/repo-webhooks/{itemID} [get]
func (h *Handler) GetRepoWebhook(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeProjectEnv)
}
