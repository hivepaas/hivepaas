package webhookdto

import (
	"net/http"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	idMaxLen = 100
)

type HandleRepoWebhookReq struct {
	Request *http.Request `json:"-"`
	ID      string        `json:"-"`
}

func NewHandleRepoWebhookReq() *HandleRepoWebhookReq {
	return &HandleRepoWebhookReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *HandleRepoWebhookReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.ID, true,
		1, idMaxLen, "id")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type HandleRepoWebhookResp struct {
	Meta *basedto.Meta `json:"meta"`
}
