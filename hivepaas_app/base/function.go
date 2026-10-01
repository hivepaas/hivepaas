package base

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
