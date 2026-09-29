package reposerverservice

import (
	"context"
	"errors"
	"io"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/reposerveragentuc"
)

// Session is a repository server an agent runs for as long as the session is
// open.
type Session struct {
	// Port is where the server listens, on the address the agent was reached by.
	Port int
	// Fingerprint is the SHA-256 fingerprint of the server's certificate.
	Fingerprint string

	cancel    context.CancelFunc
	conn      *grpc.ClientConn
	done      chan struct{}
	err       error
	closeOnce sync.Once
}

// Open asks the agent at agentAddr to run a repository server, and returns once
// it listens. The server stops when the session is closed or ctx ends.
func Open(ctx context.Context, agentAddr string, req *reposerveragentuc.RunReq) (*Session, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	session, err := openOn(client.CreateAuthCtx(ctx), conn, req)
	if err != nil {
		_ = conn.Close()
		return nil, hperrors.Wrap(err)
	}
	session.conn = conn
	return session, nil
}

// openOn opens a session on a connection; ctx carries the agent's credentials.
func openOn(ctx context.Context, conn grpc.ClientConnInterface, req *reposerveragentuc.RunReq) (*Session, error) {
	streamCtx, cancel := context.WithCancel(ctx)
	stream, err := agentproto.NewRepoServerServiceClient(conn).RepoServer(streamCtx)
	if err != nil {
		cancel()
		return nil, hperrors.Wrap(err)
	}
	err = stream.Send(&agentproto.RepoServerReq{
		RepoPath:     req.RepoPath,
		RepoPassword: req.RepoPassword,
		Username:     req.Username,
		UserPassword: req.UserPassword,
		ListenHost:   req.ListenHost,
	})
	if err != nil {
		cancel()
		return nil, hperrors.Wrap(err)
	}
	resp, err := stream.Recv()
	if err != nil {
		cancel()
		return nil, hperrors.Wrap(err)
	}

	session := &Session{
		Port:        int(resp.GetPort()),
		Fingerprint: resp.GetCertFingerprint(),
		cancel:      cancel,
		done:        make(chan struct{}),
	}
	go func() {
		var err error
		defer close(session.done)
		// Set before done closes, a panic's included, so Err reads it once Done does.
		defer func() {
			if err != nil {
				session.err = hperrors.Wrap(err)
			}
		}()
		defer safego.RecoverTo(&err)
		// The agent sends nothing more: the stream ends with the server.
		if _, recvErr := stream.Recv(); recvErr != nil && !errors.Is(recvErr, io.EOF) && streamCtx.Err() == nil {
			err = recvErr
		}
	}()
	return session, nil
}

// Done is closed when the server is gone.
func (s *Session) Done() <-chan struct{} {
	return s.done
}

// Err is why the server went, when it went by itself; nil otherwise.
func (s *Session) Err() error {
	select {
	case <-s.done:
		return s.err
	default:
		return nil
	}
}

// Close stops the server and waits for its stream to end.
func (s *Session) Close() error {
	var err error
	s.closeOnce.Do(func() {
		s.cancel()
		<-s.done
		if s.conn != nil {
			err = s.conn.Close()
		}
	})
	return hperrors.Wrap(err)
}
