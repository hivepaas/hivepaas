package secretguard

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// run drives one response through the guard and reports the panic, if any.
func run(t *testing.T, handler gin.HandlerFunc) (recovered string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				recovered, _ = r.(string)
			}
		}()
		ctx.Next()
	})
	engine.Use(Guard())
	engine.GET("/", handler)

	engine.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/", nil))
	return recovered
}

func TestGuardPanicsAndNamesWhereTheValueIs(t *testing.T) {
	got := run(t, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{
			"backend": gin.H{"ingest": gin.H{"password": base.EncryptionKeyPrefix + "Zm9v"}},
		}})
	})

	assert.Contains(t, got, "secretguard:")
	assert.Contains(t, got, "data.backend.ingest.password",
		"the path is what tells someone which Transform forgot to mask")
	assert.NotContains(t, got, "Zm9v", "the value itself must not reach a log")
}

// Both prefixes count, and the check reads them from base so a new one is
// covered without touching this package.
func TestGuardCoversEveryEncryptionPrefix(t *testing.T) {
	for _, prefix := range base.AllEncryptionPrefixes {
		got := run(t, func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"secret": prefix + "x"})
		})
		assert.Contains(t, got, "secretguard:", prefix)
	}
}

func TestGuardFindsValuesInsideArrays(t *testing.T) {
	got := run(t, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"forwards": []gin.H{
			{"name": "clean"},
			{"name": "leaky", "token": base.EncryptionSaltPrefix + "x"},
		}}})
	})

	assert.Contains(t, got, "data.forwards[1].token")
}

func TestGuardIsSilentForAMaskedResponse(t *testing.T) {
	got := run(t, func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"data": gin.H{"password": base.MaskedSecret}})
	})

	assert.Empty(t, got, "a masked secret is the correct outcome, not a finding")
}

// A log stream can hold these prefixes as content. Scanning non-JSON is how a
// guard earns a reputation for crying wolf and stops being read.
func TestGuardIgnoresNonJSONResponses(t *testing.T) {
	got := run(t, func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "app log line mentioning "+base.EncryptionKeyPrefix+"abc")
	})

	assert.Empty(t, got)
}

// The response still goes out exactly as the handler wrote it: the guard tees,
// it does not hold the body back.
func TestGuardDoesNotChangeTheResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(func(ctx *gin.Context) {
		defer func() { _ = recover() }()
		ctx.Next()
	})
	engine.Use(Guard())
	engine.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"kept": "value", "leaked": base.EncryptionKeyPrefix + "x"})
	})

	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, strings.Contains(rec.Body.String(), `"kept":"value"`), rec.Body.String())
}
