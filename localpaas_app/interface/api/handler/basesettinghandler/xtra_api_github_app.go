package basesettinghandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/githubappuc/githubappdto"
)

const (
	htmlSuccess = "<html><body><h3>Success</h3></body></html>"
)

func (h *Handler) GithubAppManifestFlowBegin(ctx *gin.Context, scopeType base.SettingScopeType) {
	var auth *basedto.Auth
	var err error

	scope := &base.SettingScope{}
	switch scopeType { //nolint:exhaustive
	case base.SettingScopeGlobal:
		auth, _, err = h.GetAuthGlobalSettings(ctx, base.ResourceTypeGithubApp, base.ActionTypeWrite, "")
	case base.SettingScopeProject:
		auth, scope.ProjectID, _, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "")
	default:
		h.RenderError(ctx, apperrors.NewUnsupported(fmt.Sprintf("Scope '%v'", scopeType)))
		return
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := githubappdto.NewBeginGithubAppManifestFlowReq()
	req.Scope = scope
	req.Type = base.SettingTypeGithubApp
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GithubAppUC.BeginGithubAppManifestFlow(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) GithubAppManifestFlowBeginCreation(ctx *gin.Context, scopeType base.SettingScopeType) {
	if scopeType == base.SettingScopeProject {
		_, err := h.ParseStringParam(ctx, "projectID")
		if err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	itemID, err := h.ParseStringParam(ctx, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := githubappdto.NewBeginGithubAppManifestFlowCreationReq()
	req.SettingID = itemID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GithubAppUC.BeginGithubAppManifestFlowCreation(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.Data(http.StatusOK, "text/html", reflectutil.UnsafeStrToBytes(resp.Data.PageContent))
}

func (h *Handler) GithubAppManifestFlowProgress(ctx *gin.Context, scopeType base.SettingScopeType) {
	if scopeType == base.SettingScopeProject {
		_, err := h.ParseStringParam(ctx, "projectID")
		if err != nil {
			h.RenderError(ctx, err)
			return
		}
	}

	itemID, err := h.ParseStringParam(ctx, "itemID")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := githubappdto.NewHandleGithubAppManifestFlowProgressReq()
	req.SettingID = itemID
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GithubAppUC.HandleGithubAppManifestFlowProgress(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	if resp.Data.RedirectURL != "" {
		ctx.Redirect(http.StatusFound, resp.Data.RedirectURL)
	}

	ctx.Data(http.StatusOK, "text/html", []byte(htmlSuccess))
}

func (h *Handler) GithubAppBeginReprovision(ctx *gin.Context, scopeType base.SettingScopeType) {
	var auth *basedto.Auth
	var err error
	var itemID string

	scope := &base.SettingScope{}
	switch scopeType { //nolint:exhaustive
	case base.SettingScopeGlobal:
		auth, itemID, err = h.GetAuthGlobalSettings(ctx, base.ResourceTypeGithubApp, base.ActionTypeWrite, "itemID")
	case base.SettingScopeProject:
		auth, scope.ProjectID, itemID, err = h.GetAuthProjectSettings(ctx, base.ActionTypeWrite, "itemID")
	default:
		h.RenderError(ctx, apperrors.NewUnsupported(fmt.Sprintf("Scope '%v'", scopeType)))
		return
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	req := githubappdto.NewBeginReprovisionGithubAppReq()
	req.Scope = scope
	req.Type = base.SettingTypeGithubApp
	req.ID = itemID
	if err = h.ParseJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.GithubAppUC.BeginReprovisionGithubApp(h.RequestCtx(ctx), auth, req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}
