package projectenvsettingshandler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	_ "github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/basicauthuc/basicauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
)

// ListGitCredentials Lists git credentials settings
// @Summary Lists git credentials settings
// @Description Lists git credentials settings
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvGitCredentials
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} gitcredentialdto.ListGitCredentialResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/git-credentials [get]
func (h *Handler) ListGitCredentials(ctx *gin.Context) {
	auth, projectID, projectEnvID, _, err := h.GetAuthProjectEnvSettings(ctx, base.ActionTypeRead, "")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := gitcredentialdto.NewListGitCredentialReq()
	req.Scope = entity.NewObjectScopeProjectEnv(projectID, projectEnvID)
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

// ListGitRepository Lists git repositories
// @Summary Lists git repositories
// @Description Lists git repositories
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvGitRepository
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "credential ID"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} gitcredentialdto.ListRepoResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/git-credentials/{itemID}/repositories [get]
func (h *Handler) ListGitRepository(ctx *gin.Context) {
	auth, projectID, projectEnvID, itemID, err := h.GetAuthProjectEnvSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := gitcredentialdto.NewListRepoReq()
	req.Scope = entity.NewObjectScopeProjectEnv(projectID, projectEnvID)
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

// ListGitBranch Lists branches of a git repository
// @Summary Lists branches of a git repository
// @Description Lists branches of a git repository
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvGitBranch
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "credential ID"
// @Param   owner query string true "repo owner (org, user)"
// @Param   repo query string true "repo name"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} gitcredentialdto.ListBranchResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/git-credentials/{itemID}/repository/branches [get]
func (h *Handler) ListGitBranch(ctx *gin.Context) {
	auth, projectID, projectEnvID, itemID, err := h.GetAuthProjectEnvSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := gitcredentialdto.NewListBranchReq()
	req.Scope = entity.NewObjectScopeProjectEnv(projectID, projectEnvID)
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GitCredentialUC.ListBranch(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// ListGitPullRequest Lists pull requests of a git repository
// @Summary Lists pull requests of a git repository
// @Description Lists pull requests of a git repository
// @Tags    project_env_settings
// @Produce json
// @Id      listProjectEnvGitPullRequest
// @Param   projectID path string true "project ID"
// @Param   projectEnv path string true "project Env"
// @Param   itemID path string true "credential ID"
// @Param   owner query string true "repo owner (org, user)"
// @Param   repo query string true "repo name"
// @Param   search query string false "`search=<target> (support *)`"
// @Param   pageOffset query int false "`pageOffset=offset`"
// @Param   pageLimit query int false "`pageLimit=limit`"
// @Param   sort query string false "`sort=[-]field1|field2...`"
// @Success 200 {object} gitcredentialdto.ListPullRequestResp
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Router  /projects/{projectID}/{projectEnv}/git-credentials/{itemID}/repository/pull-requests [get]
func (h *Handler) ListGitPullRequest(ctx *gin.Context) {
	auth, projectID, projectEnvID, itemID, err := h.GetAuthProjectEnvSettings(ctx, base.ActionTypeRead, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := gitcredentialdto.NewListPullRequestReq()
	req.Scope = entity.NewObjectScopeProjectEnv(projectID, projectEnvID)
	req.ID = itemID
	if err = h.ParseAndValidateRequest(ctx, req, &req.Paging); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GitCredentialUC.ListPullRequest(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
