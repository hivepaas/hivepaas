package main

import (
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMain runs the placeholder in a child process when asked to, on a local
// port: the test sends that process the signals swarm sends a container.
func TestMain(m *testing.M) {
	if os.Getenv("HP_PLACEHOLDER_RUN") == "1" {
		run([]string{"127.0.0.1:0"})
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

// serveLocally serves the placeholder on a local port until the test ends.
func serveLocally(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	stop := make(chan os.Signal, 1)
	done := make(chan struct{})
	time.AfterFunc(0, func() { serve([]net.Listener{listener}, stop); close(done) })
	t.Cleanup(func() {
		stop <- syscall.SIGTERM
		<-done
	})
	return "http://" + listener.Addr().String()
}

// Whoever opens the app's domain before it is deployed is told so, on any
// path, and the page is not kept: the app answers itself once deployed.
func TestEveryPathIsAnsweredWithTheWelcomePage(t *testing.T) {
	base := serveLocally(t)

	for _, path := range []string{"/", "/api/users/1?x=<b>"} {
		resp, err := http.Get(base + path) //nolint:noctx // a local test server
		if !assert.NoError(t, err) {
			return
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode, path)
		assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
		assert.Equal(t, "no-store", resp.Header.Get("Cache-Control"))
		assert.Contains(t, string(body), "not been deployed yet")
		assert.NotContains(t, string(body), "<b>", "nothing of the request is written back")
	}
}

// A port that cannot be listened on is skipped: the others still answer, and
// the placeholder does not exit for it.
func TestAPortThatCannotBeListenedOnIsSkipped(t *testing.T) {
	taken, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = taken.Close() }()

	listeners := listen([]string{taken.Addr().String(), "127.0.0.1:0"})
	defer func() {
		for _, l := range listeners {
			_ = l.Close()
		}
	}()

	assert.Len(t, listeners, 1)
}
