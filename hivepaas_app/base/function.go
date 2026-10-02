package base

import (
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

// FunctionRuntime is a function's language and its line. It names the runtime
// image the function is built on, function-runtime-<runtime>, released from the
// repository hivepaas/function-runtimes.
type FunctionRuntime string

const (
	FunctionRuntimeNode24    FunctionRuntime = "node24"
	FunctionRuntimePython313 FunctionRuntime = "python313"
	FunctionRuntimeGo127     FunctionRuntime = "go127"
)

var (
	AllFunctionRuntimes = []FunctionRuntime{FunctionRuntimeNode24, FunctionRuntimePython313, FunctionRuntimeGo127}
)

// Compiled says whether the runtime's functions are compiled, in a build image
// of their own, before they run on the runtime image.
func (r FunctionRuntime) Compiled() bool {
	switch r {
	case FunctionRuntimeGo127:
		return true
	case FunctionRuntimeNode24, FunctionRuntimePython313:
		return false
	}
	return false
}

// FunctionBuildImageKey is the key of a compiled runtime's build image among a
// release's function runtimes: function-runtime-go127-build is "go127-build".
func FunctionBuildImageKey(runtime FunctionRuntime) string {
	return string(runtime) + "-build"
}

// DefaultEntrypoint is where the runtime finds the handler when a function does
// not say: the handler's file (for Go, its package's directory) and its name.
func (r FunctionRuntime) DefaultEntrypoint() (file, handler string) {
	switch r {
	case FunctionRuntimeNode24:
		return "index.js", "default"
	case FunctionRuntimePython313:
		return "main.py", "handler"
	case FunctionRuntimeGo127:
		return ".", "Handle"
	}
	return "", ""
}

// FunctionContract is the version of the contract between a handler and its
// runtime. A runtime image's major version is its contract's.
type FunctionContract string

const (
	FunctionContractV1 FunctionContract = "v1"
)

var (
	AllFunctionContracts = []FunctionContract{FunctionContractV1}
)

// What one call of a function may take, and how much code a function carries
// inline. A function that says nothing gets the defaults, which are the
// runtime's own.
const (
	FunctionPort                  = 8080
	FunctionTimeoutDefault        = 30 * time.Second
	FunctionTimeoutMin            = time.Second
	FunctionTimeoutMax            = 15 * time.Minute
	FunctionMaxConcurrencyDefault = 16
	FunctionMaxConcurrencyMax     = 1000
	FunctionMaxBodySizeDefault    = 6 * unit.MB
	FunctionMaxBodySizeMin        = unit.KB
	FunctionMaxBodySizeMax        = 100 * unit.MB
	FunctionInlineCodeMaxSize     = unit.MB
	FunctionInlineCodeMaxFiles    = 100
	FunctionSystemPackagesMax     = 50
)

// FunctionReservedDir is the directory of a function's source that is
// HivePaaS's own: the Dockerfile it writes for the function goes there.
const FunctionReservedDir = ".hivepaas"

// FunctionPathMaxLen is the longest path of a file or a directory of a
// function's code.
const FunctionPathMaxLen = 255

var (
	// functionPathPattern is a path spelled plainly: no space, no quote, nothing a
	// Dockerfile or a shell would read as more than a name.
	functionPathPattern = regexp.MustCompile(`^[A-Za-z0-9._/-]+$`)

	// FunctionHandlerPatterns are the names each runtime can call.
	FunctionHandlerPatterns = map[FunctionRuntime]*regexp.Regexp{
		FunctionRuntimeNode24:    regexp.MustCompile(`^[A-Za-z_$][A-Za-z0-9_$]*$`),
		FunctionRuntimePython313: regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`),
		FunctionRuntimeGo127:     regexp.MustCompile(`^[A-Z][A-Za-z0-9_]*$`),
	}
	// FunctionEntrypointExts are the extensions of a handler's file, for the
	// runtimes whose entrypoint is a file. Node.js runs TypeScript by removing
	// its types as it loads a file.
	FunctionEntrypointExts = map[FunctionRuntime][]string{
		FunctionRuntimeNode24:    {".js", ".mjs", ".cjs", ".ts", ".mts", ".cts"},
		FunctionRuntimePython313: {".py"},
	}
)

// FunctionPathOK says whether p names a place inside a function: relative,
// within it, and spelled plainly. "." - the function's root - is a place for a
// directory only.
func FunctionPathOK(p string, dir bool) bool {
	if p == "" || len(p) > FunctionPathMaxLen || path.IsAbs(p) || !functionPathPattern.MatchString(p) {
		return false
	}
	if p == "." {
		return dir
	}
	for part := range strings.SplitSeq(p, "/") {
		if part == "" || part == "." || part == ".." {
			return false
		}
	}
	return true
}

// FunctionPathReserved says whether p is in the directory that is HivePaaS's.
func FunctionPathReserved(p string) bool {
	return p == FunctionReservedDir || strings.HasPrefix(p, FunctionReservedDir+"/")
}

// FunctionSystemPackagePattern is a Debian package name, with an optional
// =version: lowercase letters, digits, '+', '-' and '.', starting with a letter
// or a digit. A package reaches a RUN line of the function's Dockerfile, so
// nothing else may pass.
var FunctionSystemPackagePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]+(=[A-Za-z0-9.+~:-]+)?$`)
