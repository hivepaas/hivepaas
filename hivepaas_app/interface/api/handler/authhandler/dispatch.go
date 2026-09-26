package authhandler

import (
	"context"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// dispatchedAuthKey is unexported: only a value this package puts in a context
// carries a caller, and a request from outside the process can set its headers
// but never its context.
type dispatchedAuthKey struct{}

// WithDispatchedAuth is the context of a request the backend sends to its own
// router on behalf of a caller it has already authenticated - an MCP tool. The
// handler takes the caller from it instead of from a header, then checks access
// exactly as it would for a request from outside.
func WithDispatchedAuth(ctx context.Context, auth *basedto.Auth) context.Context {
	return context.WithValue(ctx, dispatchedAuthKey{}, auth)
}

// dispatchedAuth is the caller a dispatched request carries, or nil. It is a
// copy with nothing but the user: an access check writes the resources it
// allowed into the auth it is given, and a handler reading what an earlier
// dispatched request was allowed would filter by the wrong set.
func dispatchedAuth(ctx context.Context) *basedto.Auth {
	auth, _ := ctx.Value(dispatchedAuthKey{}).(*basedto.Auth)
	if auth == nil || auth.User == nil {
		return nil
	}
	return &basedto.Auth{User: auth.User}
}

// GetAPIKeyAuth authenticates a request by API key alone: the two HIVEPAAS-API-*
// headers, or "Authorization: Bearer <keyId>:<secret>" for clients that let a
// person set that header and no other. A session token is refused: it expires
// within minutes and would end up pasted into a config file. It answers the
// key's ID beside the auth, which carries the key's user and limits but not the
// key itself.
func (h *Handler) GetAPIKeyAuth(ctx *gin.Context) (*basedto.Auth, string, error) {
	keyID, secret, err := h.getAuthAPIKey(ctx)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	if keyID == "" {
		keyID, secret, err = bearerAPIKey(ctx.GetHeader("Authorization"))
		if err != nil {
			return nil, "", hperrors.Wrap(err)
		}
	}
	if keyID == "" {
		return nil, "", hperrors.Wrap(hperrors.ErrNoSession)
	}
	auth, err := h.sessionUC.GetCurrentAuthByAPIKey(h.RequestCtx(ctx), keyID, secret)
	if err != nil {
		return nil, "", hperrors.Wrap(err)
	}
	return auth, keyID, nil
}

// bearerAPIKey reads "Bearer <keyId>:<secret>". An empty header is no key; a
// bearer value without a colon is a session token, which is refused rather than
// taken for a key.
func bearerAPIKey(header string) (keyID, secret string, err error) {
	if header == "" {
		return "", "", nil
	}
	scheme, value, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return "", "", hperrors.Wrap(hperrors.ErrSessionAPIKeyInvalid)
	}
	keyID, secret, found = strings.Cut(strings.TrimSpace(value), ":")
	if !found || keyID == "" || secret == "" {
		return "", "", hperrors.Wrap(hperrors.ErrSessionAPIKeyInvalid)
	}
	return keyID, secret, nil
}
