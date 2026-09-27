package reposerveragentuc

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

// fakeKopia stands in for the kopia binary: each subcommand runs a shell script,
// and every call is recorded with its environment.
type fakeKopia struct {
	mu      sync.Mutex
	calls   [][]string
	envs    [][]string
	scripts map[string]string
}

const (
	readyScript = `echo "SERVER CERT SHA256: ab12cd"; echo "SERVER ADDRESS: https://10.0.1.5:40123"; exec sleep 60`
)

func (f *fakeKopia) command(args []string, env []string) *exec.Cmd {
	f.mu.Lock()
	f.calls = append(f.calls, args)
	f.envs = append(f.envs, env)
	f.mu.Unlock()
	script := "exit 0"
	for key, s := range f.scripts {
		if strings.Contains(strings.Join(args, " "), key) {
			script = s
		}
	}
	return exec.Command("sh", "-c", script)
}

func (f *fakeKopia) call(sub string) ([]string, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i, args := range f.calls {
		if strings.Contains(strings.Join(args, " "), sub) {
			return args, f.envs[i]
		}
	}
	return nil, nil
}

func newTestUC(t *testing.T, fake *fakeKopia) *UC {
	t.Helper()
	logger, err := logging.NewZapLogger(&config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return &UC{
		logger:       logger,
		command:      fake.command,
		sessionRoot:  t.TempDir(),
		memLimit:     "256MiB",
		startTimeout: 5 * time.Second,
		stopGrace:    time.Second,
	}
}

func runReq() *RunReq {
	return &RunReq{
		RepoPath: "/host/srv/repos/r1", RepoPassword: "repo-password",
		Username: "hivepaas@data-backup", UserPassword: "user-password", ListenHost: "10.0.1.5",
	}
}

// processAlive says whether a process of the given ID is still running.
func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// The server is started on the repository, its port and fingerprint handed over
// once it listens, and it is stopped, its session gone, when the caller is done.
func TestRunKeepsTheServerForTheCallersTime(t *testing.T) {
	fake := &fakeKopia{scripts: map[string]string{"server start": readyScript}}
	uc := newTestUC(t, fake)
	ctx, cancel := context.WithCancel(context.Background())

	var ready *Ready
	done := make(chan error, 1)
	go func() {
		done <- uc.Run(ctx, runReq(), func(r *Ready) error {
			ready = r
			cancel()
			return nil
		})
	}()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("Run did not return once the caller was done")
	}
	if assert.NotNil(t, ready) {
		assert.Equal(t, 40123, ready.Port)
		assert.Equal(t, "ab12cd", ready.Fingerprint)
	}

	connect, connectEnv := fake.call("repository connect")
	assert.Contains(t, connect, "filesystem")
	assert.Contains(t, connect, "--path=/host/srv/repos/r1")
	assert.Contains(t, connectEnv, "KOPIA_PASSWORD=repo-password")

	userAdd, _ := fake.call("server user add")
	assert.Contains(t, userAdd, "hivepaas@data-backup")
	assert.Contains(t, userAdd, "--user-password=user-password")

	start, startEnv := fake.call("server start")
	assert.Contains(t, start, "--address=10.0.1.5:0")
	assert.Contains(t, start, "--tls-generate-cert")
	assert.Contains(t, startEnv, "GOMEMLIMIT=256MiB")
	assert.Contains(t, startEnv, "KOPIA_PASSWORD=repo-password")
	assert.NotContains(t, strings.Join(start, " "), "--server-password", "the control password stays out of argv")

	// Every call used the session's own config file, and the session is gone.
	configFlag := start[0]
	assert.True(t, strings.HasPrefix(configFlag, "--config-file="+uc.sessionRoot), configFlag)
	entries, _ := os.ReadDir(uc.sessionRoot)
	assert.Empty(t, entries, "the session directory is removed")
}

// A user already in the repository has its password set instead.
func TestRunSetsAnExistingUsersPassword(t *testing.T) {
	fake := &fakeKopia{scripts: map[string]string{
		"server user add": `echo "error getting new user profile: hivepaas@data-backup: user already exists" >&2; exit 1`,
		"server start":    readyScript,
	}}
	uc := newTestUC(t, fake)
	ctx, cancel := context.WithCancel(context.Background())

	err := uc.Run(ctx, runReq(), func(*Ready) error { cancel(); return nil })

	assert.NoError(t, err)
	userSet, _ := fake.call("server user set")
	assert.Contains(t, userSet, "--user-password=user-password")
}

// A server that does not listen in time is stopped, and the error says what it
// printed.
func TestRunFailsWhenTheServerDoesNotStart(t *testing.T) {
	fake := &fakeKopia{scripts: map[string]string{
		"server start": `echo "error opening repository: invalid repository password"; exec sleep 60`,
	}}
	uc := newTestUC(t, fake)
	uc.startTimeout = 300 * time.Millisecond

	err := uc.Run(context.Background(), runReq(), func(*Ready) error {
		t.Error("ready must not be called")
		return nil
	})

	assert.ErrorContains(t, err, "invalid repository password")
	entries, _ := os.ReadDir(uc.sessionRoot)
	assert.Empty(t, entries)
}

// A server that exits while in use ends the session with what it printed.
func TestRunFailsWhenTheServerExits(t *testing.T) {
	fake := &fakeKopia{scripts: map[string]string{
		"server start": `echo "SERVER CERT SHA256: ab12cd"; echo "SERVER ADDRESS: https://10.0.1.5:40123"; ` +
			`sleep 0.2; echo "fatal: repository gone"; exit 1`,
	}}
	uc := newTestUC(t, fake)

	err := uc.Run(context.Background(), runReq(), func(*Ready) error { return nil })

	assert.ErrorContains(t, err, "repository gone")
}

// A repository that cannot be connected to starts no server.
func TestRunFailsWhenTheRepositoryCannotBeConnected(t *testing.T) {
	fake := &fakeKopia{scripts: map[string]string{
		"repository connect": `echo "repository not initialized in the provided storage" >&2; exit 1`,
	}}
	uc := newTestUC(t, fake)

	err := uc.Run(context.Background(), runReq(), func(*Ready) error { return nil })

	assert.ErrorContains(t, err, "not initialized")
	start, _ := fake.call("server start")
	assert.Nil(t, start)
}

// The server's whole process group is stopped: nothing it started outlives it.
func TestRunStopsTheServersProcessGroup(t *testing.T) {
	pidFile := t.TempDir() + "/child"
	fake := &fakeKopia{scripts: map[string]string{
		"server start": `sleep 60 & echo $! > ` + pidFile + `; ` + readyScript,
	}}
	uc := newTestUC(t, fake)
	ctx, cancel := context.WithCancel(context.Background())

	err := uc.Run(ctx, runReq(), func(*Ready) error { cancel(); return nil })

	assert.NoError(t, err)
	raw, readErr := os.ReadFile(pidFile)
	if assert.NoError(t, readErr) {
		var pid int
		_, _ = fmt.Sscan(strings.TrimSpace(string(raw)), &pid)
		assert.Eventually(t, func() bool { return !processAlive(pid) }, 5*time.Second, 50*time.Millisecond)
	}
}

// A ready callback that fails - the caller is gone - stops the server.
func TestRunStopsTheServerWhenReadyFails(t *testing.T) {
	fake := &fakeKopia{scripts: map[string]string{"server start": readyScript}}
	uc := newTestUC(t, fake)

	err := uc.Run(context.Background(), runReq(), func(*Ready) error { return errors.New("stream closed") })

	assert.ErrorContains(t, err, "stream closed")
}
