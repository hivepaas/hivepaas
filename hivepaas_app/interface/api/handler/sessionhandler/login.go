package sessionhandler

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tiendc/gofn"

	_ "github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
)

const (
	cookieAccessToken     = "access_token"
	cookieRefreshToken    = "refresh_token"
	cookieAccessHTTPOnly  = false
	cookieRefreshHTTPOnly = true
	cookieRefreshPath     = "/sessions/refresh" // to avoid unnecessary sending of refresh token to server
)

// LoginGetOptions Gets login options
// @Summary Gets login options
// @Description Gets login options
// @Tags    sessions
// @Produce json
// @Id      getLoginOptions
// @Success 200 {object} sessiondto.GetLoginOptionsResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /auth/login-options [get]
func (h *Handler) LoginGetOptions(ctx *gin.Context) {
	req := sessiondto.NewGetLoginOptionsReq()
	if err := h.ParseAndValidateRequest(ctx, req, nil); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.GetLoginOptions(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

// LoginWithPassword Login to system with username/password
// @Summary Login to system with username/password
// @Description When you get response's `next_step` with value `NextMfa`, you need to call the API
// @Description `/login-with-passcode` to complete the login process.
// @Tags    sessions
// @Produce json
// @Id      loginWithPassword
// @Param   body body sessiondto.LoginWithPasswordReq true "request data"
// @Success 200 {object} sessiondto.LoginWithPasswordResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /auth/login-with-password [post]
func (h *Handler) LoginWithPassword(ctx *gin.Context) {
	req := sessiondto.NewLoginWithPasswordReq()
	req.AcceptLanguage = h.ParseRequestLang(ctx)
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.LoginWithPassword(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// Writes a portion of data response to cookies
	if resp.Data.Session != nil {
		h.writeSessionDataToCookies(ctx, resp.Data.Session, true)
	}

	ctx.JSON(http.StatusOK, resp)
}

// LoginWithPasscode Login to system with passcode after using password
// @Summary Login to system with passcode after using password
// @Description Login to system with passcode after using password
// @Tags    sessions
// @Produce json
// @Id      loginWithPasscode
// @Param   body body sessiondto.LoginWithPasscodeReq true "request data"
// @Success 200 {object} sessiondto.LoginWithPasscodeResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /auth/login-with-passcode [post]
func (h *Handler) LoginWithPasscode(ctx *gin.Context) {
	req := sessiondto.NewLoginWithPasscodeReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.LoginWithPasscode(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// Writes a portion of data response to cookies
	if resp.Data.Session != nil {
		h.writeSessionDataToCookies(ctx, resp.Data.Session, true)
	}

	ctx.JSON(http.StatusOK, resp)
}

// LoginWithAPIKey Login to system with API key
// @Summary Login to system with API key
// @Description Login to system with API key
// @Tags    sessions
// @Produce json
// @Id      loginWithAPIKey
// @Param   body body sessiondto.LoginWithAPIKeyReq true "request data"
// @Success 200 {object} sessiondto.LoginWithAPIKeyResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /auth/login-with-api-key [post]
func (h *Handler) LoginWithAPIKey(ctx *gin.Context) {
	req := sessiondto.NewLoginWithAPIKeyReq()
	if err := h.ParseAndValidateJSONBody(ctx, req); err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.LoginWithAPIKey(h.RequestCtx(ctx), req)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, resp)
}

func (h *Handler) writeSessionDataToCookies(ctx *gin.Context, sessionResp *sessiondto.BaseCreateSessionResp,
	writeRefreshOnly bool) {
	ctx.SetSameSite(http.SameSiteLaxMode)

	secure := !config.Current.IsDevEnv()
	timeNow := timeutil.NowUTC()

	if !writeRefreshOnly {
		accessAge := int(sessionResp.AccessTokenExp.Sub(timeNow).Seconds())
		baseURL := strings.SplitN(config.Current.BaseURL, ".", 2) //nolint:mnd
		ctx.SetCookie(cookieAccessToken, sessionResp.AccessToken, accessAge, "",
			baseURL[len(baseURL)-1], secure, cookieAccessHTTPOnly)
	}

	refreshPath := ""
	if cookieRefreshPath != "" {
		refreshPath = gofn.Must(url.JoinPath(config.Current.HTTPServer.BasePath, cookieRefreshPath))
	}

	// Writes refresh token only (requested by FE team)
	refreshAge := int(sessionResp.RefreshTokenExp.Sub(timeNow).Seconds())
	ctx.SetCookie(cookieRefreshToken, sessionResp.RefreshToken, refreshAge, refreshPath, "",
		secure, cookieRefreshHTTPOnly)
}

func (h *Handler) clearSessionDataFromCookies(ctx *gin.Context) {
	// MaxAge<0 means delete cookie now
	ctx.SetCookie(cookieAccessToken, "", -1, "", "", false, false)
	ctx.SetCookie(cookieRefreshToken, "", -1, "", "", false, false)
}

// RefreshSession Refreshes the current user session
// @Summary Refreshes the current user session
// @Description Refreshes the current user session. Refresh token is required via either
// @Description `Authorization` header or `refresh_token` cookie.
// @Tags    sessions
// @Produce json
// @Id      refreshSession
// @Success 200 {object} sessiondto.RefreshSessionResp
// @Failure 400 {object} apperrors.ErrorInfo
// @Failure 500 {object} apperrors.ErrorInfo
// @Router  /sessions/refresh [post]
func (h *Handler) RefreshSession(ctx *gin.Context) {
	var user *basedto.User
	var err error
	// Refresh token is retrieved from either `Authorization` header or `refresh_token` cookie
	if refreshToken, _ := ctx.Cookie(cookieRefreshToken); refreshToken != "" {
		user, err = h.authHandler.GetCurrentUserByToken(ctx, refreshToken)
	} else {
		user, err = h.authHandler.GetCurrentUser(ctx)
	}
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	resp, err := h.sessionUC.RefreshSession(h.RequestCtx(ctx), user)
	if err != nil {
		h.RenderError(ctx, err)
		return
	}

	// Writes a portion of data response to cookies
	h.writeSessionDataToCookies(ctx, resp.Data.BaseCreateSessionResp, true)

	ctx.JSON(http.StatusOK, resp)
}
