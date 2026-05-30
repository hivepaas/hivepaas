package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/repowebhookuc/repowebhookdto"
)

// ListRepoWebhook Lists webhook settings
// @Summary Lists webhook settings
// @Description Lists webhook settings
// @Tags    settings
// @Produce json
// @Id      listSettingRepoWebhook
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} repowebhookdto.ListRepoWebhookResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/repo-webhooks [get]
func (h *Handler) ListRepoWebhook(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeGlobal)
}

// GetRepoWebhook Gets webhook setting details
// @Summary Gets webhook setting details
// @Description Gets webhook setting details
// @Tags    settings
// @Produce json
// @Id      getSettingRepoWebhook
// @Param   itemID path string true "setting ID"
// @Success 200 {object} repowebhookdto.GetRepoWebhookResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/repo-webhooks/{itemID} [get]
func (h *Handler) GetRepoWebhook(ctx *gin.Context) {
	h.GetSetting(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeGlobal)
}

// CreateRepoWebhook Creates a new webhook setting
// @Summary Creates a new webhook setting
// @Description Creates a new webhook setting
// @Tags    settings
// @Produce json
// @Id      createSettingRepoWebhook
// @Param   body body repowebhookdto.CreateRepoWebhookReq true "request data"
// @Success 201 {object} repowebhookdto.CreateRepoWebhookResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/repo-webhooks [post]
func (h *Handler) CreateRepoWebhook(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeGlobal)
}

// UpdateRepoWebhook Updates webhook
// @Summary Updates webhook
// @Description Updates webhook
// @Tags    settings
// @Produce json
// @Id      updateSettingRepoWebhook
// @Param   itemID path string true "setting ID"
// @Param   body body repowebhookdto.UpdateRepoWebhookReq true "request data"
// @Success 200 {object} repowebhookdto.UpdateRepoWebhookResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/repo-webhooks/{itemID} [put]
func (h *Handler) UpdateRepoWebhook(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeGlobal)
}

// UpdateRepoWebhookStatus Updates webhook status
// @Summary Updates webhook status
// @Description Updates webhook status
// @Tags    settings
// @Produce json
// @Id      updateSettingRepoWebhookStatus
// @Param   itemID path string true "setting ID"
// @Param   body body repowebhookdto.UpdateRepoWebhookStatusReq true "request data"
// @Success 200 {object} repowebhookdto.UpdateRepoWebhookStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/repo-webhooks/{itemID}/status [put]
func (h *Handler) UpdateRepoWebhookStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeGlobal)
}

// DeleteRepoWebhook Deletes webhook setting
// @Summary Deletes webhook setting
// @Description Deletes webhook setting
// @Tags    settings
// @Produce json
// @Id      deleteSettingRepoWebhook
// @Param   itemID path string true "setting ID"
// @Success 200 {object} repowebhookdto.DeleteRepoWebhookResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/repo-webhooks/{itemID} [delete]
func (h *Handler) DeleteRepoWebhook(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeRepoWebhook, base.ObjectScopeGlobal)
}
