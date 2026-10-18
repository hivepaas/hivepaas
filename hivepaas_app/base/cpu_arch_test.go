package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCPUArchFromDocker(t *testing.T) {
	for reported, want := range map[string]CPUArch{
		"x86_64": CPUArchAMD64, "amd64": CPUArchAMD64,
		"aarch64": CPUArchARM64, "arm64": CPUArchARM64,
	} {
		got, ok := CPUArchFromDocker(reported)
		assert.True(t, ok, reported)
		assert.Equal(t, want, got, reported)
	}
	for _, reported := range []string{"", "armv7l", "riscv64", "X86_64"} {
		_, ok := CPUArchFromDocker(reported)
		assert.False(t, ok, reported)
	}
}
