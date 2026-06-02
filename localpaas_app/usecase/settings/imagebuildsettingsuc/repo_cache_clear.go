package imagebuildsettingsuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/imagebuildsettingsuc/imagebuildsettingsdto"
)

func (uc *UC) ClearRepoCache(
	ctx context.Context,
	auth *basedto.Auth,
	req *imagebuildsettingsdto.ClearRepoCacheReq,
) (*imagebuildsettingsdto.ClearRepoCacheResp, error) {
	// TODO: add implementation

	return &imagebuildsettingsdto.ClearRepoCacheResp{}, nil
}
