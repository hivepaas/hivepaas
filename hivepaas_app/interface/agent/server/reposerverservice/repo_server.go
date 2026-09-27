package reposerverservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/reposerveragentuc"
)

// Runner runs a repository server: the agent's use case.
type Runner interface {
	Run(ctx context.Context, req *reposerveragentuc.RunReq, ready func(*reposerveragentuc.Ready) error) error
}

// RepoServer runs a repository server for as long as the client keeps the stream:
// until it closes its side, cancels, or is gone.
func RepoServer(runner Runner, stream agentproto.RepoServerService_RepoServerServer) error {
	req, err := stream.Recv()
	if err != nil {
		return hperrors.Wrap(err)
	}

	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()
	// The client's half-close, or anything else it sends, ends the session.
	go func() {
		_, _ = stream.Recv()
		cancel()
	}()

	err = runner.Run(ctx, &reposerveragentuc.RunReq{
		RepoPath:     req.GetRepoPath(),
		RepoPassword: req.GetRepoPassword(),
		Username:     req.GetUsername(),
		UserPassword: req.GetUserPassword(),
		ListenHost:   req.GetListenHost(),
	}, func(ready *reposerveragentuc.Ready) error {
		return stream.Send(&agentproto.RepoServerResp{
			Port:            int32(ready.Port), //nolint:gosec // a port
			CertFingerprint: ready.Fingerprint,
		})
	})
	return hperrors.Wrap(err)
}
