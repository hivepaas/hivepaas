package authhandler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/interface/api/handler"
	"github.com/localpaas/localpaas/localpaas_app/permission"
	"github.com/localpaas/localpaas/localpaas_app/usecase/sessionuc"
)

var (
	NoAccessCheck = (*permission.AccessCheck)(nil)
)

type AuthHandler struct {
	*handler.BaseHandler
	sessionUC *sessionuc.SessionUC
}

func NewAuthHandler(
	baseHandler *handler.BaseHandler,
	sessionUC *sessionuc.SessionUC,
) *AuthHandler {
	return &AuthHandler{
		BaseHandler: baseHandler,
		sessionUC:   sessionUC,
	}
}

func (h *AuthHandler) GetCurrentUser(ctx *gin.Context) (*basedto.User, error) {
	token, err := h.getAuthToken(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if token != "" {
		user, err := h.sessionUC.GetCurrentUserByJWT(h.RequestCtx(ctx), token)
		if err != nil {
			return nil, apperrors.New(err)
		}
		return user, nil
	}

	keyID, secret, err := h.getAuthAPIKey(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if keyID != "" && secret != "" {
		user, err := h.sessionUC.GetCurrentUserByAPIKey(h.RequestCtx(ctx), keyID, secret)
		if err != nil {
			return nil, apperrors.New(err)
		}
		return user, nil
	}

	return nil, apperrors.New(apperrors.ErrNoSession)
}

func (h *AuthHandler) GetCurrentUserByToken(ctx *gin.Context, token string) (*basedto.User, error) {
	user, err := h.sessionUC.GetCurrentUserByJWT(h.RequestCtx(ctx), token)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return user, nil
}

func (h *AuthHandler) GetCurrentAuth(ctx *gin.Context, accessCheck *permission.AccessCheck) (*basedto.Auth, error) {
	auth, err := h.getCurrentAuth(ctx)
	if err != nil {
		return auth, apperrors.New(err) // NOTE: on error, still return `auth`
	}

	if err = h.sessionUC.VerifyAuth(ctx, auth, accessCheck); err != nil {
		// NOTE: even on error, we still return the `auth` object so the client code
		// still can be able to check permission with another method.
		return auth, apperrors.New(err)
	}

	return auth, nil
}

func (h *AuthHandler) getCurrentAuth(ctx *gin.Context) (*basedto.Auth, error) {
	token, err := h.getAuthToken(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if token != "" {
		auth, err := h.sessionUC.GetCurrentAuthByJWT(h.RequestCtx(ctx), token)
		if err != nil {
			return auth, apperrors.New(err) // NOTE: on error, still return `auth`
		}
		return auth, nil
	}

	keyID, secret, err := h.getAuthAPIKey(ctx)
	if err != nil {
		return nil, apperrors.New(err)
	}
	if keyID != "" && secret != "" {
		auth, err := h.sessionUC.GetCurrentAuthByAPIKey(h.RequestCtx(ctx), keyID, secret)
		if err != nil {
			return auth, apperrors.New(err) // NOTE: on error, still return `auth`
		}
		return auth, nil
	}

	return nil, apperrors.New(apperrors.ErrNoSession)
}

// getAuthToken gets token from request header `Authorization`.
// The value should be in form of `Bearer <token-data>`.
func (h *AuthHandler) getAuthToken(ctx *gin.Context) (token string, err error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", nil
	}
	tokenParts := strings.SplitN(authHeader, " ", 2) //nolint:mnd
	if len(tokenParts) != 2 || tokenParts[1] == "" {
		return "", apperrors.New(apperrors.ErrSessionJWTInvalid)
	}
	return tokenParts[1], nil
}

func (h *AuthHandler) getAuthAPIKey(ctx *gin.Context) (keyID, secret string, err error) {
	keyID = ctx.GetHeader("LOCALPAAS-API-KEY-ID")
	secret = ctx.GetHeader("LOCALPAAS-API-SECRET-KEY")
	if keyID == "" && secret == "" {
		return "", "", nil
	}
	if keyID == "" || secret == "" {
		return "", "", apperrors.New(apperrors.ErrSessionAPIKeyInvalid)
	}
	return keyID, secret, nil
}
