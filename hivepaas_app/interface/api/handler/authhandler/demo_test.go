package authhandler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const demoUserID = "demo-user"

func withDemoUser(t *testing.T) {
	t.Helper()
	previous := config.Current()
	cfg := &config.Config{}
	cfg.Users.Demo.UserID = demoUserID
	config.SetCurrent(cfg)
	t.Cleanup(func() { config.SetCurrent(previous) })
}

// signedIn is a session whose token is the user's.
type signedIn struct {
	fakeSession
	user *basedto.User
}

func (s *signedIn) GetCurrentUserByJWT(context.Context, string) (*basedto.User, error) {
	return s.user, nil
}

func (s *signedIn) GetCurrentAuthByJWT(context.Context, string) (*basedto.Auth, error) {
	return &basedto.Auth{User: s.user}, nil
}

// call sends a request to route through a router whose handler authenticates
// it, and answers what the handler was given.
func call(t *testing.T, userID, method, route, path string) (*basedto.Auth, error) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := &Handler{BaseHandler: newTestHandler(nil).BaseHandler,
		sessionUC: &signedIn{user: &basedto.User{User: &entity.User{ID: userID}}}}
	var auth *basedto.Auth
	var err error
	engine := gin.New()
	engine.Handle(method, route, func(ctx *gin.Context) { auth, err = h.GetCurrentAuth(ctx, NoAccessCheck) })
	req := httptest.NewRequest(method, path, nil)
	req.Header.Set("Authorization", "Bearer token")
	engine.ServeHTTP(httptest.NewRecorder(), req)
	return auth, err
}

// The demo user reads, and does nothing else, whatever access check a handler
// asks for - none, or one that writes under a READ.
func TestTheDemoUserOnlyReads(t *testing.T) {
	withDemoUser(t)
	const app = "/api/projects/:projectID/:projectEnv/apps/:appID"

	auth, err := call(t, demoUserID, http.MethodGet, app, "/api/projects/p/e/apps/a")
	assert.NoError(t, err)
	assert.NotNil(t, auth)

	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete} {
		auth, err = call(t, demoUserID, method, app+"/deploy", "/api/projects/p/e/apps/a/deploy")
		assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized, method)
		assert.Nil(t, auth, "%s: no auth to fall back on", method)
	}
}

// Some reads act, or hand over what a public demo must keep: a shell in a
// container, its files, secrets, certificates' keys, backups.
func TestTheDemoUserCannotReadWhatActsOrExposes(t *testing.T) {
	withDemoUser(t)
	for _, route := range []string{
		"/api/projects/:projectID/:projectEnv/apps/:appID/terminal",
		"/api/projects/:projectID/:projectEnv/apps/:appID/container/file-download",
		"/api/projects/:projectID/:projectEnv/apps/:appID/container/file-upload/stream",
		"/api/projects/:projectID/:projectEnv/apps/:appID/secrets/:itemID/download",
		"/api/projects/:projectID/secrets/:itemID/download-token",
		"/api/settings/ssl-certs/:itemID/download",
		"/api/projects/:projectID/:projectEnv/backup-snapshots/:itemID/download",
		"/api/settings/github-apps/:itemID/manifest-flow/begin",
		"/api/projects/:projectID/github-apps/:itemID/manifest-flow/progress",
	} {
		_, err := call(t, demoUserID, http.MethodGet, route, route)
		assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized, route)
	}
}

// Keeping a session going, ending it, and working out a schedule's next runs
// write nothing anyone shares.
func TestTheDemoUserKeepsItsSession(t *testing.T) {
	withDemoUser(t)
	for _, c := range []struct{ method, route string }{
		{http.MethodPost, "/api/sessions/refresh"},
		{http.MethodDelete, "/api/sessions"},
		{http.MethodPost, "/api/settings/sched-jobs/calc-next-runs"},
	} {
		_, err := call(t, demoUserID, c.method, c.route, c.route)
		assert.NoError(t, err, c.method+" "+c.route)
	}
	_, err := call(t, demoUserID, http.MethodPost, "/api/sessions/delete-all", "/api/sessions/delete-all")
	assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized, "every visitor shares the demo's sessions")
}

// Nobody else is concerned.
func TestOtherUsersAreNotConcerned(t *testing.T) {
	withDemoUser(t)

	_, err := call(t, "someone", http.MethodPost, "/api/projects/:projectID", "/api/projects/p")
	assert.NoError(t, err)
	_, err = call(t, "someone", http.MethodGet, "/api/apps/:appID/terminal", "/api/apps/a/terminal")
	assert.NoError(t, err)
}

// A request an MCP tool dispatches is held to the same, by its own method.
func TestADispatchedRequestOfTheDemoUserOnlyReads(t *testing.T) {
	withDemoUser(t)
	gin.SetMode(gin.TestMode)
	h := newTestHandler(&fakeSession{})
	caller := &basedto.Auth{User: &basedto.User{User: &entity.User{ID: demoUserID}}}
	var err error
	engine := gin.New()
	engine.POST("/api/projects/:projectID", func(ctx *gin.Context) { _, err = h.GetCurrentAuth(ctx, NoAccessCheck) })
	req := httptest.NewRequestWithContext(WithDispatchedAuth(context.Background(), caller),
		http.MethodPost, "/api/projects/p", nil)

	engine.ServeHTTP(httptest.NewRecorder(), req)

	assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized)
}

// GetCurrentUser, which the session and user handlers use, is held to it too.
func TestGetCurrentUserOfTheDemoUserOnlyReads(t *testing.T) {
	withDemoUser(t)
	gin.SetMode(gin.TestMode)
	h := &Handler{BaseHandler: newTestHandler(nil).BaseHandler,
		sessionUC: &signedIn{user: &basedto.User{User: &entity.User{ID: demoUserID}}}}
	var user *basedto.User
	var err error
	engine := gin.New()
	engine.PUT("/api/users/:userID/password", func(ctx *gin.Context) { user, err = h.GetCurrentUser(ctx) })
	req := httptest.NewRequest(http.MethodPut, "/api/users/current/password", nil)
	req.Header.Set("Authorization", "Bearer token")

	engine.ServeHTTP(httptest.NewRecorder(), req)

	assert.ErrorIs(t, err, hperrors.ErrUserDemoUnauthorized)
	assert.Nil(t, user)
}
