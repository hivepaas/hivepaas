package projectsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	_ "github.com/localpaas/localpaas/localpaas_app/usecase/settings/basicauthuc/basicauthdto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
)

// ListGitCredential Lists git credential settings
// @Summary Lists git credential settings
// @Description Lists git credential settings
// @Tags    project_settings
// @Produce json
// @Id      listProjectGitCredential
// @Param   projectID path string true "project ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} gitcredentialdto.ListGitCredentialResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/git-credentials [get]
func (h *Handler) ListGitCredential(ctx *gin.Context) {
	auth, projectID, _, err := h.GetAuthProjectSettings(ctx, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := gitcredentialdto.NewListGitCredentialReq()
	req.Scope = base.NewSettingScopeProject(projectID)
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

// ListGitRepo Lists git repos
// @Summary Lists git repos
// @Description Lists git repos
// @Tags    project_settings
// @Produce json
// @Id      listProjectGitRepo
// @Param   projectID path string true "project ID"
// @Param   itemID path string true "credential ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} gitcredentialdto.ListRepoResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /projects/{projectID}/git-credentials/{itemID}/repositories [get]
func (h *Handler) ListGitRepo(ctx *gin.Context) {
	auth, projectID, itemID, err := h.GetAuthProjectSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := gitcredentialdto.NewListRepoReq()
	req.Scope = base.NewSettingScopeProject(projectID)
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GitCredentialUC.ListRepo(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
