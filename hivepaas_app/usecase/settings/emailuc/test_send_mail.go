package emailuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/emailuc/emaildto"
	"github.com/hivepaas/hivepaas/services/email"
)

func (uc *UC) TestSendMail(
	ctx context.Context,
	auth *basedto.Auth,
	req *emaildto.TestSendMailReq,
) (_ *emaildto.TestSendMailResp, err error) {
	conf := req.ToEntity()
	err = email.SendMail(ctx, conf, []string{req.TestRecipient}, req.TestSubject, req.TestContent)
	if err != nil {
		return nil, apperrors.New(err)
	}
	return &emaildto.TestSendMailResp{}, nil
}
