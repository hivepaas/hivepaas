package authhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/api/handler"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
)

// fakeSession records what the handler asks of the session use case.
type fakeSession struct {
	apiKeys  map[string]string // key id -> secret
	verified []*basedto.Auth
	jwtCalls int
	verify   error
}

func (f *fakeSession) GetCurrentUserByJWT(context.Context, string) (*basedto.User, error) {
	f.jwtCalls++
	return nil, hperrors.Wrap(hperrors.ErrSessionJWTInvalid)
}

func (f *fakeSession) GetCurrentUserByAPIKey(_ context.Context, keyID, secret string) (*basedto.User, error) {
	if f.apiKeys[keyID] != secret {
		return nil, hperrors.Wrap(hperrors.ErrAPIKeyInvalid)
	}
	return &basedto.User{}, nil
}

func (f *fakeSession) GetCurrentAuthByJWT(context.Context, string) (*basedto.Auth, error) {
	f.jwtCalls++
	return nil, hperrors.Wrap(hperrors.ErrSessionJWTInvalid)
}

func (f *fakeSession) GetCurrentAuthByAPIKey(ctx context.Context, keyID, secret string) (*basedto.Auth, error) {
	user, err := f.GetCurrentUserByAPIKey(ctx, keyID, secret)
	if err != nil {
		return nil, err
	}
	return &basedto.Auth{User: user}, nil
}

func (f *fakeSession) VerifyAuth(_ context.Context, auth *basedto.Auth, _ permission.AccessCheck) error {
	f.verified = append(f.verified, auth)
	return f.verify
}

func newTestHandler(session *fakeSession) *Handler {
	return &Handler{BaseHandler: &handler.BaseHandler{}, sessionUC: session}
}

func ginContext(ctx context.Context, headers map[string]string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/x", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	c.Request = req
	return c
}

func TestDispatchedAuthIsOnlyTakenFromTheContext(t *testing.T) {
	h := newTestHandler(&fakeSession{})
	caller := &basedto.Auth{User: &basedto.User{}}

	got, err := h.GetCurrentAuth(ginContext(WithDispatchedAuth(context.Background(), caller), nil), NoAccessCheck)
	assert.NoError(t, err)
	assert.Same(t, caller.User, got.User)

	// Nothing a request from outside can send stands for it.
	_, err = h.GetCurrentAuth(ginContext(context.Background(), map[string]string{
		"X-Dispatched-Auth": "admin", "Dispatched-Auth": "1",
	}), NoAccessCheck)
	assert.ErrorIs(t, err, hperrors.ErrNoSession)

	user, err := h.GetCurrentUser(ginContext(WithDispatchedAuth(context.Background(), caller), nil))
	assert.NoError(t, err)
	assert.Same(t, caller.User, user)
}

// The handler still checks access, on the caller, as for a request from outside.
func TestDispatchedAuthStillVerifiesAccess(t *testing.T) {
	session := &fakeSession{verify: hperrors.Wrap(hperrors.ErrUnauthorized)}
	h := newTestHandler(session)
	caller := &basedto.Auth{User: &basedto.User{}}
	check := &permission.ModuleAccessCheck{
		BaseAccessCheck: permission.BaseAccessCheck{Action: base.ActionTypeWrite},
		Module:          base.ResourceModuleCluster,
	}

	_, err := h.GetCurrentAuth(ginContext(WithDispatchedAuth(context.Background(), caller), nil), check)
	assert.ErrorIs(t, err, hperrors.ErrUnauthorized)
	if assert.Len(t, session.verified, 1) {
		assert.Same(t, caller.User, session.verified[0].User)
	}
}

// An access check writes what it allowed into the auth it is given; the next
// dispatched request of the same caller must not start from that.
func TestEachDispatchedRequestGetsAFreshAuth(t *testing.T) {
	h := newTestHandler(&fakeSession{})
	caller := &basedto.Auth{User: &basedto.User{}}
	ctx := WithDispatchedAuth(context.Background(), caller)

	first, err := h.GetCurrentAuth(ginContext(ctx, nil), NoAccessCheck)
	assert.NoError(t, err)
	first.AllowedResources = map[base.ResourceType][]string{base.ResourceTypeApp: {"app1"}}

	second, err := h.GetCurrentAuth(ginContext(ctx, nil), NoAccessCheck)
	assert.NoError(t, err)
	assert.Nil(t, second.AllowedResources)
	assert.Nil(t, caller.AllowedResources)
}

func TestGetAPIKeyAuthTakesBothForms(t *testing.T) {
	session := &fakeSession{apiKeys: map[string]string{"key1": "s3cret"}}
	h := newTestHandler(session)

	for name, headers := range map[string]map[string]string{
		"headers": {"HIVEPAAS-API-KEY-ID": "key1", "HIVEPAAS-API-SECRET-KEY": "s3cret"},
		"bearer":  {"Authorization": "Bearer key1:s3cret"},
	} {
		auth, err := h.GetAPIKeyAuth(ginContext(context.Background(), headers))
		assert.NoError(t, err, name)
		assert.NotNil(t, auth, name)
	}

	for name, headers := range map[string]map[string]string{
		"a session token": {"Authorization": "Bearer eyJhbGciOiJIUzI1NiJ9.e30.sig"},
		"a wrong secret":  {"Authorization": "Bearer key1:wrong"},
		"no secret":       {"Authorization": "Bearer key1:"},
		"basic":           {"Authorization": "Basic a2V5MTpzM2NyZXQ="},
		"nothing":         {},
	} {
		_, err := h.GetAPIKeyAuth(ginContext(context.Background(), headers))
		assert.Error(t, err, name)
	}
	assert.Zero(t, session.jwtCalls, "a session token is never tried")
}
