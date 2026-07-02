package sessionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// SSOOAuthBegin Starts OAuth SSO flow
// @Summary Starts OAuth SSO flow
// @Description Starts OAuth SSO flow
// @Tags    sessions_auth
// @Produce json
// @Id      ssoOAuthBegin
// @Success 302 "on success redirect to provider OAuth URL"
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /auth/sso/{provider} [get]
func (h *Handler) SSOOAuthBegin(ctx *gin.Context) {
	provider, err := h.ParseStringParam(ctx, "provider")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	err = h.sessionUC.InitOAuthProvider(ctx, &sessiondto.InitOAuthProviderReq{Provider: provider})
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()
	gothic.BeginAuthHandler(ctx.Writer, ctx.Request)
}

// SSOOAuthCallback Begins SSO flow
// @Summary Begins SSO flow
// @Description Begins SSO flow
// @Tags    users
// @Produce json
// @Id      ssoOAuthCallback
// @Param   provider path string true "provider name"
// @Success 302 "on success redirect to the dashboard page"
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /auth/sso/callback/{provider} [get]
// @Router  /auth/sso/callback/{provider} [post]
func (h *Handler) SSOOAuthCallback(ctx *gin.Context) {
	provider, err := h.ParseStringParam(ctx, "provider")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()

	oauthUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// Create a new session for OAuth user in our service
	sessionReq := sessiondto.NewCreateOAuthSessionReq()
	sessionReq.User = &oauthUser
	sessionResp, err := h.sessionUC.CreateOAuthSession(h.RequestCtx(ctx), sessionReq)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// Write session data to request cookies as we will redirect to a front-end path
	// and passing long tokens in the URL is risky.
	h.writeSessionDataToCookies(ctx, &sessionResp.BaseCreateSessionResp, false)

	// Redirect client to front-end page
	ctx.Redirect(http.StatusFound, config.Current.DashboardSsoSuccessURL())
}
