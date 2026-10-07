package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appsettingsuc/appsettingsdto"
)

// FillBase and FillMiddle are a request's fields embedded by pointer, two
// deep, as many request types embed theirs.
type FillBase struct {
	Name string `json:"name"`
}

type FillMiddle struct {
	*FillBase
}

type fillReq struct {
	UpdateVer int `json:"updateVer"`
	*FillMiddle
}

// Validate reads the embedded fields, as the request types' do.
func (r *fillReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		vld.StrLen(&r.Name, 1, 50).OnError(vld.SetField("name", nil)),
	))
}

// parseBody runs ParseAndValidateJSONBody on a body, as a handler does.
func parseBody(t *testing.T, body string, req any) error {
	t.Helper()
	gin.SetMode(gin.TestMode)
	h := &BaseHandler{}
	var err error
	engine := gin.New()
	engine.POST("/", func(c *gin.Context) { err = h.ParseAndValidateJSONBody(c, req) })
	engine.ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)))
	return err
}

// A body that names none of the embedded fields leaves them empty, not nil:
// refused for what is missing, not a panic.
func TestABodyWithoutTheEmbeddedFieldsIsRefusedForThem(t *testing.T) {
	req := &fillReq{}

	err := parseBody(t, `{}`, req)

	var vldErrs hperrors.ValidationErrors
	assert.ErrorAs(t, err, &vldErrs, "a validation error, naming the field")
	if assert.NotNil(t, req.FillMiddle) {
		assert.NotNil(t, req.FillBase)
	}
}

// The request that panicked on {}: its embedded fields are filled, and it
// parses.
func TestAnEmptyBodyUpdatesNoEnvVarsRatherThanPanicking(t *testing.T) {
	req := appsettingsdto.NewUpdateAppEnvVarsReq()

	assert.NotPanics(t, func() { _ = parseBody(t, `{}`, req) })
	assert.NotNil(t, req.AppEnvVarsBaseReq)
}
