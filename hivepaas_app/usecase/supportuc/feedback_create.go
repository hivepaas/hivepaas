package supportuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/supportuc/supportdto"
)

func (uc *UC) CreateFeedback(
	ctx context.Context,
	auth *basedto.Auth,
	req *supportdto.CreateFeedbackReq,
) (*supportdto.CreateFeedbackResp, error) {
	// TODO: add implementation
	return &supportdto.CreateFeedbackResp{}, nil
}
