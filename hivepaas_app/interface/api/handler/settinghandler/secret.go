package settinghandler

import (
	"github.com/gin-gonic/gin"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
)

// ListSecret Lists secrets
// @Summary Lists secrets
// @Description Lists secrets
// @Tags    settings
// @Produce json
// @Id      listSettingSecrets
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} secretdto.ListSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/secrets [get]
func (h *Handler) ListSecret(ctx *gin.Context) {
	h.ListSetting(ctx, base.ResourceTypeSecret, base.ObjectScopeGlobal)
}

// CreateSecret Creates a new secret
// @Summary Creates a new secret
// @Description Creates a new secret
// @Tags    settings
// @Produce json
// @Id      createSettingSecret
// @Param   body body secretdto.CreateSecretReq true "request data"
// @Success 201 {object} secretdto.CreateSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/secrets [post]
func (h *Handler) CreateSecret(ctx *gin.Context) {
	h.CreateSetting(ctx, base.ResourceTypeSecret, base.ObjectScopeGlobal)
}

// UpdateSecret Updates secret
// @Summary Updates secret
// @Description Updates secret
// @Tags    settings
// @Produce json
// @Id      updateSettingSecret
// @Param   itemID path string true "setting ID"
// @Param   body body secretdto.UpdateSecretReq true "request data"
// @Success 201 {object} secretdto.UpdateSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/secrets/{itemID} [put]
func (h *Handler) UpdateSecret(ctx *gin.Context) {
	h.UpdateSetting(ctx, base.ResourceTypeSecret, base.ObjectScopeGlobal)
}

// UpdateSecretStatus Updates secret status
// @Summary Updates secret status
// @Description Updates secret status
// @Tags    settings
// @Produce json
// @Id      updateSettingSecretStatus
// @Param   itemID path string true "setting ID"
// @Param   body body secretdto.UpdateSecretStatusReq true "request data"
// @Success 201 {object} secretdto.UpdateSecretStatusResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/secrets/{itemID}/status [put]
func (h *Handler) UpdateSecretStatus(ctx *gin.Context) {
	h.UpdateSettingStatus(ctx, base.ResourceTypeSecret, base.ObjectScopeGlobal)
}

// DeleteSecret Deletes a secret
// @Summary Deletes a secret
// @Description Deletes a secret
// @Tags    settings
// @Produce json
// @Id      deleteSettingSecret
// @Param   itemID path string true "setting ID"
// @Success 200 {object} secretdto.DeleteSecretResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/secrets/{itemID} [delete]
func (h *Handler) DeleteSecret(ctx *gin.Context) {
	h.DeleteSetting(ctx, base.ResourceTypeSecret, base.ObjectScopeGlobal)
}
