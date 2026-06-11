package executil

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunPipeline(t *testing.T) {
	t.Run("Empty commands", func(t *testing.T) {
		out, err := RunPipeline()
		assert.NoError(t, err)
		assert.Empty(t, out)
	})

	t.Run("Single command success", func(t *testing.T) {
		cmd := exec.Command("echo", "hello")
		out, err := RunPipeline(cmd)
		assert.NoError(t, err)
		assert.Equal(t, "hello\n", out)
	})

	t.Run("Single command failure", func(t *testing.T) {
		cmd := exec.Command("nonexistent_command_123_xyz")
		out, err := RunPipeline(cmd)
		assert.Error(t, err)
		assert.Empty(t, out)
	})

	t.Run("Two commands pipeline success", func(t *testing.T) {
		cmd1 := exec.Command("echo", "hello world")
		cmd2 := exec.Command("grep", "world")
		out, err := RunPipeline(cmd1, cmd2)
		assert.NoError(t, err)
		assert.Equal(t, "hello world\n", out)
	})

	t.Run("Three commands pipeline success", func(t *testing.T) {
		cmd1 := exec.Command("echo", "line1\nline2\nline3")
		cmd2 := exec.Command("grep", "line")
		cmd3 := exec.Command("wc", "-l")
		out, err := RunPipeline(cmd1, cmd2, cmd3)
		assert.NoError(t, err)
		// wc -l output might have leading whitespaces on some OS (e.g. macOS vs Linux)
		assert.Equal(t, "3", strings.TrimSpace(out))
	})

	t.Run("Pipeline intermediate command failure", func(t *testing.T) {
		cmd1 := exec.Command("echo", "hello")
		cmd2 := exec.Command("nonexistent_command_123_xyz")
		cmd3 := exec.Command("grep", "hello")
		out, err := RunPipeline(cmd1, cmd2, cmd3)
		assert.Error(t, err)
		assert.Empty(t, out)
	})

	t.Run("Pipeline last command failure", func(t *testing.T) {
		cmd1 := exec.Command("echo", "hello")
		cmd2 := exec.Command("sh", "-c", "exit 42")
		out, err := RunPipeline(cmd1, cmd2)
		assert.Error(t, err)
		assert.Empty(t, out)
	})

	t.Run("Combined stdout and stderr on error", func(t *testing.T) {
		cmd1 := exec.Command("sh", "-c", "echo 'stdout message'; echo 'error message' >&2; exit 1")
		out, err := RunPipeline(cmd1)
		assert.Error(t, err)
		assert.Contains(t, out, "stdout message")
		assert.Contains(t, out, "error message")
	})

	t.Run("Piped stdout and stderr on error", func(t *testing.T) {
		cmd1 := exec.Command("sh", "-c", "echo 'stage1 error' >&2; echo 'stage1 stdout'")
		cmd2 := exec.Command("sh", "-c", "read input; echo 'stage2 error' >&2; echo \"stage2: $input\"; exit 1")
		out, err := RunPipeline(cmd1, cmd2)
		assert.Error(t, err)
		assert.Contains(t, out, "stage1 error")
		assert.Contains(t, out, "stage2 error")
		assert.Contains(t, out, "stage2: stage1 stdout")
	})
}
