package base_test

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Each release names, for every function runtime, the image its functions are
// built on, pinned by digest. A runtime that compiles its functions names its
// build image as well.
func TestEachReleaseNamesAnImageForEveryFunctionRuntime(t *testing.T) {
	pinned := regexp.MustCompile(`^ghcr\.io/hivepaas/function-runtime-[a-z0-9-]+:\d+\.\d+\.\d+@sha256:[0-9a-f]{64}$`)
	for name, release := range map[string]*base.ReleaseInfo{"stable": base.StableVersion, "beta": base.BetaVersion} {
		want := map[string]bool{}
		for _, runtime := range base.AllFunctionRuntimes {
			want[string(runtime)] = true
			if runtime.Compiled() {
				want[base.FunctionBuildImageKey(runtime)] = true
			}
		}
		assert.Len(t, release.FunctionRuntimes, len(want), name)
		for key, image := range release.FunctionRuntimes {
			assert.True(t, want[key], "%s names a runtime that does not exist: %s", name, key)
			assert.Regexp(t, pinned, image, name)
			assert.Contains(t, image, "/function-runtime-"+key+":", name)
		}
	}
}

func TestOnlyGoIsCompiled(t *testing.T) {
	assert.True(t, base.FunctionRuntimeGo127.Compiled())
	assert.False(t, base.FunctionRuntimeNode24.Compiled())
	assert.False(t, base.FunctionRuntimePython313.Compiled())
	assert.Equal(t, "go127-build", base.FunctionBuildImageKey(base.FunctionRuntimeGo127))
}
