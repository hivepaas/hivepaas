// Package imagescanagentuc is what an agent does for image scanning: a scan's
// container, on its node.
package imagescanagentuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
	"github.com/hivepaas/hivepaas/services/docker"
)

type UC struct {
	runner *imagescanrun.Runner
}

func New(dockerManager docker.Manager) *UC {
	return &UC{runner: &imagescanrun.Runner{Docker: dockerManager}}
}

// Run scans an image on this node.
func (uc *UC) Run(ctx context.Context, req *imagescanrun.RunReq) (resp *imagescanrun.RunResp, err error) {
	// The agent's calls have no recovery of their own: a panic here would take
	// the agent down with every build and run on it.
	defer func() {
		if r := recover(); r != nil {
			resp, err = nil, hperrors.NewPanic(r)
		}
	}()
	resp, err = uc.runner.Run(ctx, req)
	return resp, hperrors.Wrap(err)
}
