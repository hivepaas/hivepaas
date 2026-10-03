package main

import (
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMain runs main in a child process when asked to: the test sends that
// process the signals swarm sends a container.
func TestMain(m *testing.M) {
	if os.Getenv("HP_PLACEHOLDER_RUN") == "1" {
		main()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// start runs the placeholder, and gives it a moment to be waiting.
func start(t *testing.T) *exec.Cmd {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^$")
	cmd.Env = append(os.Environ(), "HP_PLACEHOLDER_RUN=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill() })
	time.Sleep(200 * time.Millisecond)
	return cmd
}

// waitExit is the process's exit code, or -1 when it has not exited within 3s,
// and is killed then.
func waitExit(cmd *exec.Cmd) int {
	var timedOut atomic.Bool
	timer := time.AfterFunc(3*time.Second, func() {
		timedOut.Store(true)
		_ = cmd.Process.Kill()
	})
	_ = cmd.Wait()
	timer.Stop()
	if timedOut.Load() {
		return -1
	}
	return cmd.ProcessState.ExitCode()
}

// It keeps running: a task that exits is restarted by swarm, again and again.
func TestThePlaceholderKeepsRunning(t *testing.T) {
	cmd := start(t)

	assert.Equal(t, -1, waitExit(cmd), "it is still running")
}

// It stops at once, and cleanly, when the container is stopped (SIGTERM) or
// interrupted (SIGINT): a deployment replacing it does not wait for the grace
// period, and the task is not counted as failed.
func TestThePlaceholderStopsCleanlyOnASignal(t *testing.T) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT} {
		cmd := start(t)

		assert.NoError(t, cmd.Process.Signal(sig))
		assert.Equal(t, 0, waitExit(cmd), sig.String())
	}
}
