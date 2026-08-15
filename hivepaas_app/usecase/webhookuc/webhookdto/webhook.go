package webhookdto

import (
	"net/http"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
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
func (req *HandleRepoWebhookReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.ID, true,
		1, idMaxLen, "id")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type HandleRepoWebhookResp struct {
	Meta *basedto.Meta `json:"meta"`
}
