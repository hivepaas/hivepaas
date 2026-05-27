package schedjobdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

const (
	calcNextSchedulesMaxCount = 10
)

type CalcNextRunsReq struct {
	*ScheduleReq
	Count int `json:"count"`
}

func NewCalcNextRunsReq() *CalcNextRunsReq {
	return &CalcNextRunsReq{}
}

func (req *CalcNextRunsReq) ModifyRequest() error {
	req.CronExpr = strings.TrimSpace(req.CronExpr)
	if req.InitialTime.IsZero() {
		req.InitialTime = timeutil.NowUTC()
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *CalcNextRunsReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, basedto.ValidateNumber(&req.Count, true,
		1, calcNextSchedulesMaxCount, "count")...)
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type CalcNextRunsResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data []time.Time   `json:"data"`
}
