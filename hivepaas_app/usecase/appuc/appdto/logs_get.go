package appdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	// DefaultAppLogsTail bounds a response that asked for no bound.
	//
	// Docker streams whatever the time range holds, so without a tail a range
	// of a few days on a talkative app is one unbounded response. Every read
	// gets a ceiling, and a caller who wants more asks for more.
	DefaultAppLogsTail = 1000

	// MaxAppLogsTail is the most one response may carry. It matches the cap on
	// stored logs, which is the paged reader for anything larger.
	MaxAppLogsTail = MaxLogHistoryLimit
)

type GetAppLogsReq struct {
	ProjectID    string            `json:"-"`
	ProjectEnvID string            `json:"-"`
	AppID        string            `json:"-"`
	TaskID       string            `json:"-" mapstructure:"taskId"`
	Follow       bool              `json:"-" mapstructure:"follow"`
	Since        time.Time         `json:"-" mapstructure:"since"`
	Duration     timeutil.Duration `json:"-" mapstructure:"duration"`
	Tail         int               `json:"-" mapstructure:"tail"`
	Timestamps   *bool             `json:"-" mapstructure:"timestamps"`
}

func NewGetAppLogsReq() *GetAppLogsReq {
	return &GetAppLogsReq{}
}

// ApplyDefaults bounds what was left unbounded.
func (req *GetAppLogsReq) ApplyDefaults() {
	if req.Tail <= 0 {
		req.Tail = DefaultAppLogsTail
	}
}

func (req *GetAppLogsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateNumber(&req.Tail, false, 0, MaxAppLogsTail, "tail")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppLogsResp struct {
	Meta *basedto.Meta    `json:"meta"`
	Data *AppLogsDataResp `json:"data"`
}

type AppLogsDataResp struct {
	StaticLogs       []*tasklog.LogFrame        `json:"logs"`
	LogsStream       <-chan []*tasklog.LogFrame `json:"-"`
	LogsStreamCloser func() error               `json:"-"`
}
