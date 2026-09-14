// Package secretguard panics, in development only, when a response carries a
// value that is plainly still encrypted.
//
// It catches one mistake: a DTO that forgot to mask an entity.EncryptedField,
// which is a whole field leaking on every call and is invisible to a reviewer
// who does not already know the field exists. It caught nothing else, and must
// not be read as saying a response is free of secrets - a password a user typed
// into an env var carries no marker at all. auditdetail.scrub makes the same
// point about the same prefixes: this is a backstop, and a backstop that is
// mistaken for the mechanism is worse than none.
package secretguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// maxCaptureBytes bounds what one response costs us. A body larger than this is
// scanned as far as it goes: a leak in the first megabyte is still a leak, and
// holding an entire log page in memory to look for one is not worth it.
const maxCaptureBytes = 1 << 20

// Guard returns the middleware. Register it only in a development environment -
// it copies every JSON response body, and it panics.
func Guard() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		capture := &capturingWriter{ResponseWriter: ctx.Writer, body: &bytes.Buffer{}}
		ctx.Writer = capture

		ctx.Next()

		// Only JSON. A log stream or a file download may hold these prefixes as
		// content rather than as a leak, and scanning those is how a guard
		// earns a reputation for crying wolf.
		if !strings.Contains(ctx.Writer.Header().Get("Content-Type"), "application/json") {
			return
		}

		found := encryptedPaths(capture.body.Bytes())
		if len(found) == 0 {
			return
		}
		panic(fmt.Sprintf(
			"secretguard: %s %s returned %d still-encrypted value(s) at %s - "+
				"a DTO is handing out an EncryptedField instead of basedto.MaskedSecret",
			ctx.Request.Method, ctx.Request.URL.Path, len(found), strings.Join(found, ", ")))
	}
}

// encryptedPaths walks the body and reports where each encrypted value sits.
//
// It reports the path rather than the value: "data.backend.ingest.password" is
// what tells someone which Transform forgot to mask, and the ciphertext itself
// would only put another copy of the secret into a log.
func encryptedPaths(body []byte) []string {
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil // not JSON after all, or truncated at the capture limit
	}
	var found []string
	walk("", decoded, &found)
	return found
}

func walk(path string, node any, found *[]string) {
	switch value := node.(type) {
	case string:
		for _, prefix := range base.AllEncryptionPrefixes {
			if strings.HasPrefix(value, prefix) {
				*found = append(*found, path)
				return
			}
		}
	case map[string]any:
		for key, child := range value {
			walk(join(path, key), child, found)
		}
	case []any:
		for i, child := range value {
			walk(fmt.Sprintf("%s[%d]", path, i), child, found)
		}
	}
}

func join(path, key string) string {
	if path == "" {
		return key
	}
	return path + "." + key
}

// capturingWriter tees the body so the response still goes out as written. The
// panic then happens after the client already has it, which is loud without
// changing what any endpoint returns.
type capturingWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// The wrapped writer's error is passed through unchanged: gin is the caller and
// expects exactly what it would have got without this wrapper in between.
func (w *capturingWriter) Write(data []byte) (int, error) {
	w.capture(data)
	return w.ResponseWriter.Write(data) //nolint:wrapcheck
}

func (w *capturingWriter) WriteString(s string) (int, error) {
	w.capture([]byte(s))
	return w.ResponseWriter.WriteString(s) //nolint:wrapcheck
}

func (w *capturingWriter) capture(data []byte) {
	if remaining := maxCaptureBytes - w.body.Len(); remaining > 0 {
		w.body.Write(data[:min(remaining, len(data))])
	}
}

var _ http.ResponseWriter = (*capturingWriter)(nil)
