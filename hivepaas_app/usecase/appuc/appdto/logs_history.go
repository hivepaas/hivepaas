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
// Levels and Streams are comma-separated. There is no field for query text:
// see services/logging/victorialogs/query.go.
type GetAppLogHistoryReq struct {
	ProjectID    string    `json:"-"`
	ProjectEnvID string    `json:"-"`
	AppID        string    `json:"-"`
	Start        time.Time `json:"-" mapstructure:"start"`
	End          time.Time `json:"-" mapstructure:"end"`
	Limit        int       `json:"-" mapstructure:"limit"`
	Search       string    `json:"-" mapstructure:"search"`
	Levels       string    `json:"-" mapstructure:"levels"`
	Streams      string    `json:"-" mapstructure:"streams"`
}

func NewGetAppLogHistoryReq() *GetAppLogHistoryReq {
	return &GetAppLogHistoryReq{}
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

func (req *GetAppLogHistoryReq) LevelList() []string  { return splitList(req.Levels) }
func (req *GetAppLogHistoryReq) StreamList() []string { return splitList(req.Streams) }

func splitList(s string) []string {
	out := []string{}
	for _, p := range strings.Split(s, ",") {
		if p = strings.ToLower(strings.TrimSpace(p)); p != "" {
			out = append(out, p)
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
	validators = append(validators, basedto.ValidateSlice(req.LevelList(), false, 0, logHistoryLevels, "levels")...)
	validators = append(validators, basedto.ValidateSlice(req.StreamList(), false, 0, logHistoryStreams, "streams")...)
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
