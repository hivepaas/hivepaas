package sessionhandler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"github.com/tiendc/gofn"

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
// @Param   trustedDeviceId query string false "the device to trust, so no passcode is asked on it for a while"
// @Success 302 "redirect to the provider; on a failure, to the sign-in page with ssoError, the error's code"
// @Security
// @Router  /auth/sso/{provider} [get]
func (h *Handler) SSOOAuthBegin(ctx *gin.Context) {
	provider, err := h.ParseStringParam(ctx, "provider")
	if err != nil {
		h.redirectSSOError(ctx, err)
		return
	}

	req := sessiondto.NewInitOAuthProviderReq()
	req.Provider = provider
	if err = h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.redirectSSOError(ctx, err)
		return
	}

	err = h.sessionUC.InitOAuthProvider(ctx, req)
	if err != nil {
		h.redirectSSOError(ctx, err)
		return
	}

	h.writeSSOTrustedDevice(ctx, ctx.Query("trustedDeviceId"))

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
// @Success 302 "redirect to the dashboard; on a failure, to the sign-in page with ssoError, the error's code"
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
// @Success 302 "redirect to the dashboard; on a failure, to the sign-in page with ssoError, the error's code"
// @Security
// @Router  /auth/sso/callback/{provider} [get]
func (h *Handler) SSOOAuthCallback(ctx *gin.Context) {
	provider, err := h.ParseStringParam(ctx, "provider")
	if err != nil {
		h.redirectSSOError(ctx, err)
		return
	}

	q := ctx.Request.URL.Query()
	q.Add("provider", provider)
	ctx.Request.URL.RawQuery = q.Encode()

	// An unknown provider, a missing or mismatched state, a code the provider
	// refuses: the sign-in is not completed, whichever it was.
	oauthUser, err := gothic.CompleteUserAuth(ctx.Writer, ctx.Request)
	if err != nil {
		h.redirectSSOError(ctx, hperrors.Wrap(hperrors.ErrUnauthorized).WithCause(err))
		return
	}

	// Create a new session for OAuth user in our service
	sessionReq := sessiondto.NewCreateOAuthSessionReq()
	sessionReq.User = &oauthUser
	sessionReq.TrustedDeviceID = h.takeSSOTrustedDevice(ctx)
	if err = h.ParseAndValidateRequest(ctx, sessionReq, nil); err != nil {
		h.redirectSSOError(ctx, err)
		return
	}
	sessionResp, err := h.sessionUC.CreateOAuthSession(h.RequestCtx(ctx), sessionReq)
	if err != nil {
		h.redirectSSOError(ctx, err)
		return
	}

	// A user with an authenticator app owes the passcode: the dashboard's sign-in asks for it with this token.
	if sessionResp.MFAToken != "" {
		h.writeSSOMFAToken(ctx, sessionResp.MFAToken)
		ctx.Redirect(http.StatusFound, config.Current().DashboardSsoSuccessURL())
		return
	}

	// Write session data to request cookies as we will redirect to a front-end path
	// and passing long tokens in the URL is risky.
	h.writeSessionDataToCookies(ctx, &sessionResp.BaseCreateSessionResp, false)

	// Redirect client to front-end page
	ctx.Redirect(http.StatusFound, config.Current().DashboardSsoSuccessURL())
}

// redirectSSOError sends the browser back to the sign-in page, which says why the sign-in failed. Every step of a
// sign-in with a provider is a page the browser opened, and an error written as JSON would leave the person on it.
func (h *Handler) redirectSSOError(ctx *gin.Context, err error) {
	errInfo := h.ErrorInfoOf(ctx, err)
	ctx.Redirect(http.StatusFound, config.Current().DashboardSsoErrorURL(errInfo.Code))
}

const (
	// cookieSSOTrustedDevice carries the device to trust from the start of a sign-in with a provider to its end,
	// only to the SSO endpoints.
	cookieSSOTrustedDevice = "sso_trusted_device"
	// cookieSSOMFAToken hands the dashboard the token its passcode step signs in with; the dashboard reads it, as
	// it reads the access token after a sign-in with a provider.
	cookieSSOMFAToken = "mfa_token"

	// ssoRoundTripMaxAge is how long a sign-in with a provider may take, there and back.
	ssoRoundTripMaxAge = 10 * 60
	// ssoMFATokenMaxAge is how long the dashboard has to pick the token up. The token runs out on its own, in a
	// minute: the cookie only carries it.
	ssoMFATokenMaxAge = 2 * 60
	// maxTrustedDeviceIDLen bounds what is kept of a device ID; one longer is not kept at all.
	maxTrustedDeviceIDLen = 200
)

func (h *Handler) ssoCookiePath() string {
	return gofn.Must(url.JoinPath(config.Current().HTTPServer.BasePath, "auth/sso"))
}

// writeSSOTrustedDevice keeps the device the sign-in page asked to trust for the provider's callback.
func (h *Handler) writeSSOTrustedDevice(ctx *gin.Context, deviceID string) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" || len(deviceID) > maxTrustedDeviceIDLen {
		return
	}
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(cookieSSOTrustedDevice, deviceID, ssoRoundTripMaxAge, h.ssoCookiePath(), "",
		!config.Current().IsDevEnv(), true)
}

// takeSSOTrustedDevice is the device kept by writeSSOTrustedDevice, which it forgets: it is for this sign-in only.
func (h *Handler) takeSSOTrustedDevice(ctx *gin.Context) string {
	deviceID, err := ctx.Cookie(cookieSSOTrustedDevice)
	if err != nil {
		return ""
	}
	ctx.SetCookie(cookieSSOTrustedDevice, "", -1, h.ssoCookiePath(), "", !config.Current().IsDevEnv(), true)
	return deviceID
}

// writeSSOMFAToken hands the passcode step's token to the dashboard, beside where the access token goes.
func (h *Handler) writeSSOMFAToken(ctx *gin.Context, mfaToken string) {
	ctx.SetSameSite(http.SameSiteLaxMode)
	baseURL := strings.SplitN(config.Current().BaseURL(), ".", 2) //nolint:mnd
	ctx.SetCookie(cookieSSOMFAToken, mfaToken, ssoMFATokenMaxAge, "",
		baseURL[len(baseURL)-1], !config.Current().IsDevEnv(), cookieAccessHTTPOnly)
}
