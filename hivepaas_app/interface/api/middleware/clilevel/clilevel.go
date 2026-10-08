// Package clilevel keeps a HivePaaS CLI from writing through an API newer than
// the one it was built for.
//
// The CLI changes an object as the dashboard does: it reads it, changes what it
// was asked to, and writes it back. A CLI built before the server added a field
// to that object would write it back without the field, and erase it. So every
// request of the CLI says which API level it was built for, and a write from a
// CLI below the server's level is refused with 426 Upgrade Required. A read is
// served: an older CLI loses nothing by reading, and a pipeline that follows its
// logs keeps working through a server upgrade.
//
// Requests that do not say they come from the CLI - the dashboard, a script -
// are not concerned.
package clilevel

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// HeaderCLI is what the CLI sends with every request: its version, and the
	// API level of the spec it was built from - "0.1.0; api-level=14".
	HeaderCLI = "HivePaaS-CLI"
	// HeaderAPILevel answers a CLI's request with the server's level, which the
	// CLI compares with its own to tell its user to update before a write fails.
	HeaderAPILevel = "HivePaaS-API-Level"

	levelPrefix = "api-level="
)

// Check refuses a write from a CLI built for an API level below level. render
// answers the error as the API answers every error.
func Check(level int, render func(*gin.Context, error)) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cli := ctx.GetHeader(HeaderCLI)
		if cli == "" {
			ctx.Next()
			return
		}
		ctx.Header(HeaderAPILevel, strconv.Itoa(level))
		if isRead(ctx.Request.Method) {
			ctx.Next()
			return
		}
		cliLevel, ok := parseLevel(cli)
		if ok && cliLevel >= level {
			ctx.Next()
			return
		}
		cliLevelText := "unknown"
		if ok {
			cliLevelText = strconv.Itoa(cliLevel)
		}
		render(ctx, hperrors.Wrap(hperrors.ErrCLIOutdated).
			WithParam("ServerLevel", level).WithParam("CLILevel", cliLevelText))
		ctx.Abort()
	}
}

func isRead(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}

// parseLevel reads the api-level of a HivePaaS-CLI header. A header without one,
// or with one that is not a number, is a CLI that cannot be trusted to write.
func parseLevel(header string) (int, bool) {
	for _, part := range strings.Split(header, ";") {
		value, found := strings.CutPrefix(strings.TrimSpace(part), levelPrefix)
		if !found {
			continue
		}
		level, err := strconv.Atoi(value)
		return level, err == nil
	}
	return 0, false
}
