package clilevel

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func serve(t *testing.T, method, cliHeader string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(Check(14, func(ctx *gin.Context, err error) {
		info, _ := hperrors.ParseError(err, translation.LangEn)
		ctx.JSON(info.Status, info)
	}))
	engine.Handle(method, "/apps", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
	req := httptest.NewRequest(method, "/apps", nil)
	if cliHeader != "" {
		req.Header.Set(HeaderCLI, cliHeader)
	}
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestAWriteFromAnOlderCLIIsRefused(t *testing.T) {
	rec := serve(t, http.MethodPut, "0.1.0; api-level=13")

	assert.Equal(t, http.StatusUpgradeRequired, rec.Code)
	assert.Equal(t, "14", rec.Header().Get(HeaderAPILevel))
	assert.Contains(t, rec.Body.String(), "ERR_CLI_OUTDATED")
	assert.Contains(t, rec.Body.String(), "the server is at API level 14, the CLI at 13")
}

func TestWhatCheckLetsThrough(t *testing.T) {
	for name, tc := range map[string]struct {
		method, header string
		want           int
	}{
		"a read from an older CLI":  {http.MethodGet, "0.1.0; api-level=13", http.StatusOK},
		"a write from a CLI as new": {http.MethodPost, "0.2.0; api-level=14", http.StatusOK},
		"a write from a newer CLI":  {http.MethodDelete, "api-level=15; 0.3.0", http.StatusOK},
		"a write from no CLI":       {http.MethodPost, "", http.StatusOK},
		"a write with no level":     {http.MethodPost, "0.1.0", http.StatusUpgradeRequired},
		"a write with a bad level":  {http.MethodPatch, "0.1.0; api-level=x", http.StatusUpgradeRequired},
	} {
		t.Run(name, func(t *testing.T) {
			rec := serve(t, tc.method, tc.header)
			assert.Equal(t, tc.want, rec.Code, rec.Body.String())
			assert.Equal(t, tc.header != "", strings.TrimSpace(rec.Header().Get(HeaderAPILevel)) != "",
				"the level is answered to the CLI, and only to it")
		})
	}
}
