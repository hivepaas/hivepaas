package emailuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/emailuc/emaildto"
	"github.com/localpaas/localpaas/services/email"
)

func (uc *UC) TestSendMail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.TestSendMailReq,
) (_ *emaildto.TestSendMailResp, err error) {
	conf := req.ToEntity()
	err = email.SendMail(ctx, conf, []string{req.TestRecipient}, req.TestSubject, req.TestContent)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return &emaildto.TestSendMailResp{}, nil
}
