package base

import (
	"regexp"
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

// FunctionSystemPackagePattern is a Debian package name, with an optional
// =version: lowercase letters, digits, '+', '-' and '.', starting with a letter
// or a digit. A package reaches a RUN line of the function's Dockerfile, so
// nothing else may pass.
var FunctionSystemPackagePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9+.-]+(=[A-Za-z0-9.+~:-]+)?$`)
