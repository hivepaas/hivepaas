package client

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
)

// A call that carries no deadline of its own is given the default.
func TestADefaultTimeoutBoundsACallThatCarriesNoDeadline(t *testing.T) {
	var left time.Duration
	err := defaultCallTimeout(42*time.Second)(context.Background(), "/m", nil, nil, nil,
		func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			deadline, ok := ctx.Deadline()
			assert.True(t, ok, "the call reached the agent unbounded")
			left = time.Until(deadline)
			return nil
		})

	assert.NoError(t, err)
	assert.InDelta(t, 42, left.Seconds(), 5)
}

// A caller that bounded its own call keeps its bound: the default is a floor for
// calls nobody bounded, not a ceiling imposed on the ones that are.
func TestACallersOwnDeadlineIsKept(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	var left time.Duration
	err := defaultCallTimeout(42*time.Second)(ctx, "/m", nil, nil, nil,
		func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
			deadline, _ := ctx.Deadline()
			left = time.Until(deadline)
			return nil
		})

	assert.NoError(t, err)
	assert.InDelta(t, time.Hour.Seconds(), left.Seconds(), 5)
}

// slowAgent answers Ping only after a delay, or not at all if the call is cut off
// first. A plain listener, not bufconn: a checkout with a vendor directory builds
// only what the code imports.
type slowAgent struct {
	agentproto.UnimplementedAgentServiceServer
	delay time.Duration
}

func (a *slowAgent) Ping(ctx context.Context, _ *agentproto.PingReq) (*agentproto.PingResp, error) {
	select {
	case <-time.After(a.delay):
		return &agentproto.PingResp{Message: "pong"}, nil
	case <-ctx.Done():
		return nil, status.Error(codes.Canceled, "the caller went away")
	}
}

func serveSlowAgent(t *testing.T, delay time.Duration) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	agentproto.RegisterAgentServiceServer(srv, &slowAgent{delay: delay})
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(srv.Stop)
	return listener.Addr().String()
}

// Dialed with a default bound, a call nobody bounded ends at the bound rather
// than waiting on an agent that is taking longer than the work can take.
func TestADialsDefaultBoundEndsAnUnboundedCall(t *testing.T) {
	addr := serveSlowAgent(t, 5*time.Second)

	conn, err := Dial(addr, WithDefaultCallTimeout(100*time.Millisecond))
	if !assert.NoError(t, err) {
		return
	}
	t.Cleanup(func() { _ = conn.Close() })

	start := time.Now()
	_, err = NewAgentServiceClient(conn).Ping(context.Background(), &PingReq{Message: "ping"})

	assert.Equal(t, codes.DeadlineExceeded, status.Code(err))
	assert.Less(t, time.Since(start), 2*time.Second, "the call waited past its bound")
}

// Dialed without one - which is how ten of the eleven agent services dial, their
// calls being as long as the work the node was asked to do - the same call runs
// to its answer. Adding the shared Dial changed nothing for them.
func TestWithoutADefaultBoundACallRunsToItsAnswer(t *testing.T) {
	addr := serveSlowAgent(t, 300*time.Millisecond)

	conn, err := Dial(addr)
	if !assert.NoError(t, err) {
		return
	}
	t.Cleanup(func() { _ = conn.Close() })

	resp, err := NewAgentServiceClient(conn).Ping(context.Background(), &PingReq{Message: "ping"})

	if assert.NoError(t, err) {
		assert.Equal(t, "pong", resp.Message)
	}
}

// The keepalive is on every connection, whether or not a bound was asked for:
// it, not a deadline, is what ends a call to an agent that went away.
func TestEveryConnectionIsKeptAlive(t *testing.T) {
	assert.Greater(t, keepaliveTime, 5*time.Second,
		"the agent's MinTime is 5s and ends a connection that pings faster")
	assert.Positive(t, keepaliveTimeout)
	assert.Less(t, keepaliveTimeout, keepaliveTime)
}

// Dial reaches an agent that is not there without blocking: grpc.NewClient is
// lazy, so a dead node shows up as an error on the call, not on the dial.
func TestDialDoesNotWaitForTheAgent(t *testing.T) {
	conn, err := Dial("127.0.0.1:1")
	if !assert.NoError(t, err) {
		return
	}
	t.Cleanup(func() { _ = conn.Close() })

	_, err = NewAgentServiceClient(conn).Ping(context.Background(), &PingReq{Message: "ping"})
	assert.Equal(t, codes.Unavailable, status.Code(err))
}
