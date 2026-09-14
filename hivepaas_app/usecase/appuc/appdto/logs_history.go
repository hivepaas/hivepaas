package appdto

import (
	"strings"
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

const (
	DefaultLogHistoryLimit  = 500
	MaxLogHistoryLimit      = 5000
	maxLogHistorySearchLen  = 256
	defaultLogHistoryWindow = time.Hour
)

var (
	logHistoryLevels  = []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	logHistoryStreams = []string{"stdout", "stderr"}
)

// GetAppLogHistoryReq is a search over an app's stored logs.
//
// Levels and Streams arrive comma-separated and are split by the decoder's
// StringToSliceHookFunc, so one value and many look the same here. Search is a value, never query text:
// see services/logging/victorialogs/query.go. Regex widens how that value is
// read, not where it is placed, so the scope still cannot be escaped.
type GetAppLogHistoryReq struct {
	ProjectID    string    `json:"-"`
	ProjectEnvID string    `json:"-"`
	AppID        string    `json:"-"`
	Start        time.Time `json:"-" mapstructure:"start"`
	End          time.Time `json:"-" mapstructure:"end"`
	Limit        int       `json:"-" mapstructure:"limit"`
	Search       string    `json:"-" mapstructure:"search"`
	// Regex reads Search as a regular expression instead of plain text. Plain
	// is the default because it is the filter the backend can answer from its
	// index; a regular expression is read row by row.
	Regex bool `json:"-" mapstructure:"regex"`
	// MatchCase compares Search case-sensitively, in either mode.
	MatchCase bool     `json:"-" mapstructure:"matchCase"`
	Levels    []string `json:"-" mapstructure:"levels"`
	Streams   []string `json:"-" mapstructure:"streams"`
}

func NewGetAppLogHistoryReq() *GetAppLogHistoryReq {
	return &GetAppLogHistoryReq{}
}

// ModifyRequest implements interface basedto.ReqModifier.
//
// It folds the lists to the spelling the allowed values are written in, so that
// `levels=ERROR, warn` is the request it obviously means rather than two
// validation errors about capitalization and a stray space. The handler calls
// this before Validate, so everything downstream - the check included - reads
// one spelling.
func (req *GetAppLogHistoryReq) ModifyRequest() error {
	req.Levels = normalizeList(req.Levels)
	req.Streams = normalizeList(req.Streams)
	return nil
}

// ApplyDefaults fills what was not asked: the last hour, up to now.
func (req *GetAppLogHistoryReq) ApplyDefaults(now time.Time) {
	if req.End.IsZero() {
		req.End = now
	}
	if req.Start.IsZero() {
		req.Start = req.End.Add(-defaultLogHistoryWindow)
	}
	if req.Limit == 0 {
		req.Limit = DefaultLogHistoryLimit
	}
}

func normalizeList(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.ToLower(strings.TrimSpace(v)); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func (req *GetAppLogHistoryReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateNumber(&req.Limit, false, 0, MaxLogHistoryLimit, "limit")...)
	validators = append(validators, basedto.ValidateStr(&req.Search, false, 0, maxLogHistorySearchLen, "search")...)
	validators = append(validators, basedto.ValidateSlice(req.Levels, false, 0, logHistoryLevels, "levels")...)
	validators = append(validators, basedto.ValidateSlice(req.Streams, false, 0, logHistoryStreams, "streams")...)
	validators = append(validators, basedto.ValidateCond(
		req.Start.IsZero() || req.End.IsZero() || req.Start.Before(req.End), "end")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppLogHistoryResp struct {
	Meta *basedto.Meta          `json:"meta"`
	Data *AppLogHistoryDataResp `json:"data"`
}

type AppLogHistoryDataResp struct {
	// Logs are oldest first.
	Logs      []*tasklog.LogFrame `json:"logs"`
	Truncated bool                `json:"truncated"`
	// NextEnd is the `end` that returns the page before this one. Set only
	// when Truncated.
	NextEnd *time.Time `json:"nextEnd,omitempty"`
}
