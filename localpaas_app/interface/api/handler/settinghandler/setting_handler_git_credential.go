package settinghandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
)

// ListGitCredential Lists git credential settings
// @Summary Lists git credential settings
// @Description Lists git credential settings
// @Tags    settings
// @Produce json
// @Id      listSettingGitCredential
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} gitcredentialdto.ListGitCredentialResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /settings/git-credentials [get]
func (h *SettingHandler) ListGitCredential(ctx *gin.Context) {
	auth, _, err := h.GetAuthGlobalSettings(ctx, "", base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := gitcredentialdto.NewListGitCredentialReq()
	req.Scope = base.NewSettingScopeGlobal()
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GitCredentialUC.ListGitCredential(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
