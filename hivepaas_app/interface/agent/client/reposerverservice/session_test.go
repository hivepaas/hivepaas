package reposerverservice

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	serverside "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/reposerverservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/reposerveragentuc"
)

// fakeRunner stands in for the agent's use case: it reports ready, then holds
// the server until its context ends, or fails as told.
type fakeRunner struct {
	req        *reposerveragentuc.RunReq
	failBefore error
	failAfter  error
	stopped    chan struct{}
}

func (f *fakeRunner) Run(ctx context.Context, req *reposerveragentuc.RunReq,
	ready func(*reposerveragentuc.Ready) error) error {
	f.req = req
	defer close(f.stopped)
	if f.failBefore != nil {
		return f.failBefore
	}
	if err := ready(&reposerveragentuc.Ready{Port: 40123, Fingerprint: "ab12cd"}); err != nil {
		return err
	}
	if f.failAfter != nil {
		return f.failAfter
	}
	<-ctx.Done()
	return nil
}

type testServer struct {
	agentproto.UnimplementedRepoServerServiceServer
	runner serverside.Runner
}

func (s *testServer) RepoServer(stream agentproto.RepoServerService_RepoServerServer) error {
	return serverside.RepoServer(s.runner, stream)
}

// dial serves the RPC over an in-memory connection.
func dial(t *testing.T, runner serverside.Runner) *grpc.ClientConn {
	t.Helper()
	listener := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	agentproto.RegisterRepoServerServiceServer(srv, &testServer{runner: runner})
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(srv.Stop)
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return listener.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}

func openReq() *reposerveragentuc.RunReq {
	return &reposerveragentuc.RunReq{
		RepoPath: "/host/srv/r1", RepoPassword: "repo-password",
		Username: "hivepaas@data-backup", UserPassword: "user-password", ListenHost: "10.0.1.5",
	}
}

// A session opens once the server listens, and closing it stops the server.
func TestSessionLivesUntilClosed(t *testing.T) {
	runner := &fakeRunner{stopped: make(chan struct{})}

	session, err := openOn(context.Background(), dial(t, runner), openReq())

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, 40123, session.Port)
	assert.Equal(t, "ab12cd", session.Fingerprint)
	assert.Equal(t, openReq(), runner.req)

	assert.NoError(t, session.Close())
	select {
	case <-runner.stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("closing the session did not stop the server")
	}
}

// The caller's context ending stops the server too.
func TestSessionEndsWithTheCallersContext(t *testing.T) {
	runner := &fakeRunner{stopped: make(chan struct{})}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := openOn(ctx, dial(t, runner), openReq())
	if !assert.NoError(t, err) {
		return
	}
	cancel()

	select {
	case <-runner.stopped:
	case <-time.After(5 * time.Second):
		t.Fatal("the caller's context did not stop the server")
	}
	_ = session.Close()
}

// A server that does not start fails the opening, with its reason.
func TestSessionOpenFailsWithTheServersReason(t *testing.T) {
	runner := &fakeRunner{stopped: make(chan struct{}), failBefore: errors.New("invalid repository password")}

	_, err := openOn(context.Background(), dial(t, runner), openReq())

	assert.ErrorContains(t, err, "invalid repository password")
}

// A server that exits while in use says so through the session.
func TestSessionReportsAServerThatExits(t *testing.T) {
	runner := &fakeRunner{stopped: make(chan struct{}), failAfter: errors.New("repository gone")}

	session, err := openOn(context.Background(), dial(t, runner), openReq())
	if !assert.NoError(t, err) {
		return
	}

	select {
	case <-session.Done():
		assert.ErrorContains(t, session.Err(), "repository gone")
	case <-time.After(5 * time.Second):
		t.Fatal("the session did not end")
	}
	_ = session.Close()
}
