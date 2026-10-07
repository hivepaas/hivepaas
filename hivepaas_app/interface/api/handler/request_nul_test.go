package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appuc/appdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/sessionuc/sessiondto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/useruc/userdto"
)

// In JSON a NUL is the escape \u0000 - a raw one is no JSON - and only when the
// backslash before it is not itself escaped.
func TestJSONHasNUL(t *testing.T) {
	for body, want := range map[string]bool{
		`{"note":"a\u0000b"}`:           true,
		`{"k\u0000":1}`:                 true,
		`{"note":"a\\u0000b"}`:          false,
		`{"note":"a\\\u0000b"}`:         true,
		`{"note":"a\\\\u0000b"}`:        false,
		`{"note":"\u00001"}`:            true,
		`{"note":"a\u0001b \u1000 \n"}`: false,
		`{}`:                            false,
		``:                              false,
	} {
		assert.Equal(t, want, jsonHasNUL([]byte(body)), body)
	}
}

// parseQuery runs ParseAndValidateRequest on a query string, as a handler does.
func parseQuery(t *testing.T, query string, req any) error {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := &BaseHandler{}
	var err error
	engine := gin.New()
	engine.GET("/", func(c *gin.Context) { err = h.ParseAndValidateRequest(c, req, nil) })
	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/?"+query, nil))
	return err
}

// A NUL cannot be stored in a text column, and in jsonb it turns into the six
// characters \u0000: a request holding one is refused, a 400 that says so,
// rather than failing as it is written or being stored as something else.
func TestARequestHoldingANULIsRefused(t *testing.T) {
	err := parseBody(t, `{"updateVer":1,"name":"web","note":"bad\u0000note"}`, appdto.NewUpdateAppReq())
	assert.ErrorIs(t, err, hperrors.ErrRequestHasNUL)
	assert.ErrorIs(t, err, hperrors.ErrBadRequest)

	assert.ErrorIs(t, parseQuery(t, "search=a%00b", &struct {
		Search string `json:"-" mapstructure:"search"`
	}{}), hperrors.ErrRequestHasNUL)

	assert.NotErrorIs(t, parseBody(t, `{"updateVer":1,"name":"web","note":"a\\u0000b"}`, appdto.NewUpdateAppReq()),
		hperrors.ErrRequestHasNUL, "an escaped backslash, then text: no NUL")
}

// What these carry is hashed, encrypted or passed on as it is, never written to
// a text column: a NUL in them goes on as it did.
func TestTheRequestsThatMayHoldANUL(t *testing.T) {
	for name, req := range map[string]any{
		"function test run": &appdto.TestRunFunctionReq{},
		"secret, created":   &secretdto.CreateSecretReq{},
		"secret, updated":   secretdto.NewUpdateSecretReq(),
		"sign in":           &sessiondto.LoginWithPasswordReq{},
		"sign up":           &userdto.CompleteUserSignupReq{},
		"password reset":    &userdto.ResetPasswordReq{},
		"password update":   userdto.NewUpdatePasswordReq(),
	} {
		err := parseBody(t, `{"password":"a\u0000b","value":"a\u0000b","request":{"body":"a\u0000b"}}`, req)
		assert.NotErrorIs(t, err, hperrors.ErrRequestHasNUL, name)
	}
}
