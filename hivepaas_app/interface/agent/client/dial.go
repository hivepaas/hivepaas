package client

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// keepaliveTime is how long a connection may go silent before the client
	// pings the agent. It has to stay above the agent's own MinTime, the shortest
	// interval between client pings it accepts: ping faster than that and the
	// agent ends the connection with ENHANCE_YOUR_CALM.
	//
	// Pinging matters because an agent that went away does not always say so. A
	// node that lost power, a container the daemon killed, a connection an
	// overlay network dropped - each leaves the app holding a connection that
	// looks open, on which a call waits for an answer that is never coming. The
	// ping is what turns that wait into an error.
	keepaliveTime = 30 * time.Second

	// keepaliveTimeout is how long a ping has to be answered before the
	// connection is taken as dead and every call on it ends.
	keepaliveTimeout = 10 * time.Second
)

// dialSettings is what Dial's options build up.
type dialSettings struct {
	defaultCallTimeout time.Duration
}

// Option configures a connection to an agent.
type Option func(*dialSettings)

// WithDefaultCallTimeout bounds the unary calls on this connection whose caller
// left them unbounded. A call that carries a deadline of its own keeps it.
//
// Pass it only where every unary method on the service finishes inside the
// duration given. Most of the agent's services have at least one method that
// does not: a command run on a node, an image scan, a node cleanup and a volume
// removal are all as long as the work the node was asked to do, and a default
// short enough to be useful on one of them would cut the others off mid-way.
// Those bound their calls at the caller, which is the only place that knows the
// bound. The keepalive above, not a deadline, is what ends their calls when the
// agent itself is gone.
func WithDefaultCallTimeout(timeout time.Duration) Option {
	return func(s *dialSettings) { s.defaultCallTimeout = timeout }
}

// Dial opens a connection to one node's agent.
//
// Every connection to an agent is made here, so that the keepalive, the
// credentials and the interceptors are the same on all of them: there were
// eleven copies of this call before, and nothing could be added to them all
// without editing eleven files.
//
// The connection is not established yet when this returns - grpc.NewClient is
// lazy - so a dead agent shows up as an error on the first call, not here.
func Dial(agentAddr string, opts ...Option) (*grpc.ClientConn, error) {
	var settings dialSettings
	for _, opt := range opts {
		opt(&settings)
	}

	dialOpts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    keepaliveTime,
			Timeout: keepaliveTimeout,
			// Ping an idle connection too. A call that is waiting for its first
			// answer holds no active stream yet, and that is exactly the wait
			// this is here to end.
			PermitWithoutStream: true,
		}),
	}
	if settings.defaultCallTimeout > 0 {
		dialOpts = append(dialOpts,
			grpc.WithChainUnaryInterceptor(defaultCallTimeout(settings.defaultCallTimeout)))
	}

	conn, err := grpc.NewClient(agentAddr, dialOpts...)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return conn, nil
}

// defaultCallTimeout gives a unary call a deadline when its caller gave it none.
// Streams are left alone: their life is the file transfer, the build or the exec
// session they carry, and only the caller driving one knows when it is over.
func defaultCallTimeout(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if _, ok := ctx.Deadline(); !ok {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...) //nolint:wrapcheck
	}
}
