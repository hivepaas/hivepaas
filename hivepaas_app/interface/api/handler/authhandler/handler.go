package authhandler

import (
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc"
)

var (
	NoAccessCheck = (*permission.AccessCheck)(nil)
)

type Handler struct {
	*handler.BaseHandler
	sessionUC *sessionuc.UC
}

func New(
	baseHandler *handler.BaseHandler,
	sessionUC *sessionuc.UC,
) *Handler {
	return &Handler{
		BaseHandler: baseHandler,
		sessionUC:   sessionUC,
	}
}

func (h *Handler) GetCurrentUser(ctx *gin.Context) (*basedto.User, error) {
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

func (h *Handler) GetCurrentUserByToken(ctx *gin.Context, token string) (*basedto.User, error) {
	user, err := h.sessionUC.GetCurrentUserByJWT(h.RequestCtx(ctx), token)
	if err != nil {
		return nil, apperrors.New(err)
	}

	return user, nil
}

func (h *Handler) GetCurrentAuth(ctx *gin.Context, accessCheck *permission.AccessCheck) (*basedto.Auth, error) {
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

func (h *Handler) getCurrentAuth(ctx *gin.Context) (*basedto.Auth, error) {
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
func (h *Handler) getAuthToken(ctx *gin.Context) (token string, err error) {
	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" {
		tokenParts := strings.SplitN(authHeader, " ", 2) //nolint:mnd
		if len(tokenParts) != 2 || tokenParts[1] == "" {
			return "", apperrors.New(apperrors.ErrSessionJWTInvalid)
		}
		return tokenParts[1], nil
	}

	wsProtoHeader := ctx.GetHeader("Sec-WebSocket-Protocol")
	if wsProtoHeader != "" {
		// Header has format: "some_proto, access_token, <token>"
		parts := strings.Split(wsProtoHeader, ",")
		for i := 0; i < len(parts)-1; i++ {
			key := strings.TrimSpace(parts[i])
			if key == "access_token" {
				return strings.TrimSpace(parts[i+1]), nil // The token is the next element
			}
		}
	}

	return "", nil
}

func (h *Handler) getAuthAPIKey(ctx *gin.Context) (keyID, secret string, err error) {
	keyID = ctx.GetHeader("HIVEPAAS-API-KEY-ID")
	secret = ctx.GetHeader("HIVEPAAS-API-SECRET-KEY")
	if keyID == "" && secret == "" {
		return "", "", nil
	}
	if keyID == "" || secret == "" {
		return "", "", apperrors.New(apperrors.ErrSessionAPIKeyInvalid)
	}
	return keyID, secret, nil
}
