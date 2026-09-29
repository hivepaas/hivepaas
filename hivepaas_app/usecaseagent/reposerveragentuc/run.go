package reposerveragentuc

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/envutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
)

const (
	// outputTailLines is how much of what kopia printed an error carries.
	outputTailLines = 10
	sessionDirMode  = 0o700
)

var (
	errServerNotListening = errors.New("the repository server did not listen")
	errServerExited       = errors.New("the repository server exited")
	errStepFailed         = errors.New("kopia failed")
)

var (
	fingerprintLine = regexp.MustCompile(`SERVER CERT SHA256: ([0-9a-f]+)`)
	addressLine     = regexp.MustCompile(`SERVER ADDRESS: https?://.*:(\d+)\s*$`)
)

// RunReq is a server to run: the repository, as the agent sees it, and the user
// its clients log in as.
type RunReq struct {
	RepoPath     string
	RepoPassword string
	Username     string
	UserPassword string
	// ListenHost is the address the server listens on: the agent's own, as the
	// caller reached it.
	ListenHost string
}

// Ready is a server listening: its port, and its certificate's SHA-256
// fingerprint, which clients pin.
type Ready struct {
	Port        int
	Fingerprint string
}

// Run runs a kopia repository server for the repository and calls ready once it
// listens. The server lives until ctx is done - the caller is finished, or gone -
// or it exits by itself, which is an error. Its session directory goes with it.
func (uc *UC) Run(ctx context.Context, req *RunReq, ready func(*Ready) error) error {
	if req == nil || req.RepoPath == "" || req.Username == "" || req.ListenHost == "" {
		return hperrors.Wrap(hperrors.ErrBadRequest).WithExtraDetail("repository, user and address are required")
	}
	memLimit, startTimeout := uc.settings()

	sessionDir, err := uc.newSessionDir()
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() {
		if err := os.RemoveAll(sessionDir); err != nil {
			uc.logger.Errorf("Failed to remove the repository server session %s: %v", sessionDir, err)
		}
	}()

	configFlag := "--config-file=" + filepath.Join(sessionDir, "repository.config")
	env := append(envutil.SafeEnviron(),
		"KOPIA_PASSWORD="+req.RepoPassword,
		"KOPIA_CHECK_FOR_UPDATES=false",
		"KOPIA_LOG_DIR="+filepath.Join(sessionDir, "logs"),
	)

	if err = uc.runStep(ctx, env, configFlag, "repository", "connect", "filesystem", "--path="+req.RepoPath,
		"--cache-directory="+filepath.Join(sessionDir, "cache"), "--no-check-for-updates"); err != nil {
		return hperrors.Wrap(err)
	}
	if err = uc.setUser(ctx, env, configFlag, req); err != nil {
		return hperrors.Wrap(err)
	}

	controlPassword, err := randomHex()
	if err != nil {
		return hperrors.Wrap(err)
	}
	serverEnv := append(append([]string{}, env...), "GOMEMLIMIT="+memLimit, "KOPIA_SERVER_PASSWORD="+controlPassword)
	server := uc.command([]string{configFlag, "server", "start",
		"--address=" + req.ListenHost + ":0",
		"--tls-generate-cert",
		"--tls-cert-file=" + filepath.Join(sessionDir, "cert.pem"),
		"--tls-key-file=" + filepath.Join(sessionDir, "key.pem"),
		"--server-username=hivepaas-control",
	}, serverEnv)
	return uc.serve(ctx, server, startTimeout, ready)
}

// serve starts the server, waits for it to listen, and keeps it until ctx is
// done or it exits.
func (uc *UC) serve(ctx context.Context, server *exec.Cmd, startTimeout time.Duration, ready func(*Ready) error) error {
	output := newServerOutput()
	server.Stdout, server.Stderr = output, output
	// Its own process group: stopping it stops whatever it started.
	server.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := server.Start(); err != nil {
		return hperrors.Wrap(err)
	}
	exited := make(chan error, 1)
	go func() {
		defer safego.RecoverWithLogger(uc.logger, "reposerver.wait")
		err := server.Wait()
		// The server's output ends with it: its reader goes too.
		output.close()
		exited <- err
	}()
	defer uc.stop(server, exited)

	timer := time.NewTimer(startTimeout)
	defer timer.Stop()
	select {
	case r := <-output.ready:
		if err := ready(r); err != nil {
			return hperrors.Wrap(err)
		}
	case err := <-exited:
		exited <- err
		return hperrors.Wrap(fmt.Errorf("%w before it listened: %s", errServerExited, output.tail()))
	case <-timer.C:
		return hperrors.Wrap(fmt.Errorf("%w within %s: %s", errServerNotListening, startTimeout, output.tail()))
	case <-ctx.Done():
		return nil
	}

	select {
	case <-ctx.Done():
		return nil
	case err := <-exited:
		exited <- err
		return hperrors.Wrap(fmt.Errorf("%w: %s", errServerExited, output.tail()))
	}
}

// stop asks the server's process group to stop, and kills it when it does not.
func (uc *UC) stop(server *exec.Cmd, exited chan error) {
	select {
	case <-exited:
		return
	default:
	}
	pgid := server.Process.Pid
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	select {
	case <-exited:
	case <-time.After(uc.stopGrace):
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
		<-exited
	}
	// Whatever it started and left running goes too.
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
}

// setUser gives the user its password: added, or set when it exists.
func (uc *UC) setUser(ctx context.Context, env []string, configFlag string, req *RunReq) error {
	err := uc.runStep(ctx, env, configFlag, "server", "user", "add", req.Username,
		"--user-password="+req.UserPassword)
	if err == nil || !strings.Contains(err.Error(), "already exists") {
		return err
	}
	return uc.runStep(ctx, env, configFlag, "server", "user", "set", req.Username,
		"--user-password="+req.UserPassword)
}

// runStep runs a kopia command to its end; its error carries what it printed.
func (uc *UC) runStep(ctx context.Context, env []string, args ...string) error {
	cmd := uc.command(args, env)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	if err := cmd.Start(); err != nil {
		return hperrors.Wrap(err)
	}
	done := make(chan error, 1)
	go func() {
		var err error
		defer func() { done <- err }()
		defer safego.RecoverTo(&err)
		err = cmd.Wait()
	}()
	select {
	case err := <-done:
		if err != nil {
			// The user's password is in the arguments: the error names the step only.
			return hperrors.Wrap(fmt.Errorf("%w: kopia %s: %w: %s", errStepFailed, strings.Join(args[1:3], " "), err,
				strings.TrimSpace(out.String())))
		}
		return nil
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		<-done
		return hperrors.Wrap(ctx.Err())
	}
}

func (uc *UC) newSessionDir() (string, error) {
	if err := os.MkdirAll(uc.sessionRoot, sessionDirMode); err != nil {
		return "", hperrors.Wrap(err)
	}
	dir, err := os.MkdirTemp(uc.sessionRoot, "session-")
	return dir, hperrors.Wrap(err)
}

func randomHex() (string, error) {
	b := make([]byte, 24) //nolint:mnd
	if _, err := rand.Read(b); err != nil {
		return "", hperrors.Wrap(err)
	}
	return hex.EncodeToString(b), nil
}

// serverOutput reads what the server prints: the lines that say it listens, and
// the last lines, for an error.
type serverOutput struct {
	io.Writer
	writer *io.PipeWriter
	ready  chan *Ready

	mu    sync.Mutex
	lines []string
}

func newServerOutput() *serverOutput {
	reader, writer := io.Pipe()
	o := &serverOutput{Writer: writer, writer: writer, ready: make(chan *Ready, 1)}
	go o.scan(reader)
	return o
}

func (o *serverOutput) scan(reader io.Reader) {
	// Whatever stops the scan - the end, a line too long for the scanner, a
	// panic - the rest is drained, so the server never blocks writing to a pipe
	// nobody reads.
	defer func() { _, _ = io.Copy(io.Discard, reader) }()
	defer safego.Recover("reposerver.output")
	scanner := bufio.NewScanner(reader)
	var fingerprint string
	port := 0
	sent := false
	for scanner.Scan() {
		line := scanner.Text()
		o.mu.Lock()
		o.lines = append(o.lines, line)
		if len(o.lines) > outputTailLines {
			o.lines = o.lines[1:]
		}
		o.mu.Unlock()
		if m := fingerprintLine.FindStringSubmatch(line); m != nil {
			fingerprint = m[1]
		}
		if m := addressLine.FindStringSubmatch(line); m != nil {
			port, _ = strconv.Atoi(m[1])
		}
		if !sent && fingerprint != "" && port != 0 {
			o.ready <- &Ready{Port: port, Fingerprint: fingerprint}
			sent = true
		}
	}
}

func (o *serverOutput) close() {
	_ = o.writer.Close()
}

func (o *serverOutput) tail() string {
	o.mu.Lock()
	defer o.mu.Unlock()
	if len(o.lines) == 0 {
		return "it printed nothing"
	}
	return strings.Join(o.lines, "; ")
}
