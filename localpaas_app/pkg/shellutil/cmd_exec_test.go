package shellutil

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cmd := exec.Command("echo", "hello world")
		var res CmdRes
		Execute(cmd, &res)
		assert.True(t, res.Succeed())
		assert.Equal(t, 0, res.Exit)
		assert.Contains(t, res.RawOutput[0], "hello world")
		assert.Contains(t, res.Output, "hello world")
	})

	t.Run("Failure", func(t *testing.T) {
		// Using a non-existent command
		cmd := exec.Command("nonexistentcommand_abc_123")
		var res CmdRes
		Execute(cmd, &res)
		assert.False(t, res.Succeed())
		assert.Error(t, res.Error)
		assert.NotEmpty(t, res.ErrorStr())
	})

	t.Run("Exit Code", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "exit 42")
		var res CmdRes
		Execute(cmd, &res)
		assert.False(t, res.Succeed())
		assert.Equal(t, 42, res.Exit)
	})
}
