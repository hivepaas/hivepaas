package sessionhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

// SSOOAuthBegin Starts OAuth SSO flow
// @Summary Starts OAuth SSO flow
// @Description Starts OAuth SSO flow
// @Tags    Sessions
// @Produce json
// @Id      ssoOAuthBegin
// @Param   provider path string true "provider name"
// @Success 302 "on success redirect to provider OAuth URL"
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
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

// SSOOAuthCallbackPost Completes the SSO flow, for a provider that posts its answer
// @Summary Completes the SSO flow, for a provider that posts its answer
// @Description The same as the GET, for a provider that returns with a form post (response_mode=form_post).
// @Tags    Sessions
// @Produce json
// @Id      ssoOAuthCallbackPost
// @Param   provider path string true "provider name"
// @Success 302 "on success redirect to the dashboard page"
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /auth/sso/callback/{provider} [post]
//
// A handler of its own, not a second @Router on SSOOAuthCallback: an operation
// id names one operation, and a client generated from the spec needs one each.
func (h *Handler) SSOOAuthCallbackPost(ctx *gin.Context) {
	h.SSOOAuthCallback(ctx)
}

// SSOOAuthCallback Completes the SSO flow
// @Summary Completes the SSO flow
// @Description The provider sends the browser back here once the person has signed in.
// @Tags    Sessions
// @Produce json
// @Id      ssoOAuthCallback
// @Param   provider path string true "provider name"
// @Success 302 "on success redirect to the dashboard page"
// @Failure 400 {object} hperrors.ErrorInfo
// @Failure 500 {object} hperrors.ErrorInfo
// @Security
// @Router  /auth/sso/callback/{provider} [get]
func (h *Handler) SSOOAuthCallback(ctx *gin.Context) {
	provider, err := h.ParseStringParam(ctx, "provider")
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()

	// An unknown provider, a missing or mismatched state, a code the provider
	// refuses: the sign-in is not completed, whichever it was.
	oauthUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		h.RenderError(ctx, hperrors.Wrap(hperrors.ErrUnauthorized).WithCause(err))
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
	ctx.Redirect(http.StatusFound, config.Current().DashboardSsoSuccessURL())
}
