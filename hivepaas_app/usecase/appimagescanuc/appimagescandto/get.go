package appimagescandto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type GetAppImageScanReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppImageScanReq() *GetAppImageScanReq {
	return &GetAppImageScanReq{}
}

func (req *GetAppImageScanReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppImageScanResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *AppImageScanResp `json:"data"`
}

// AppImageScanResp is what an app's images carry: the app's Security tab.
type AppImageScanResp struct {
	// Enabled says image scanning is switched on.
	Enabled bool `json:"enabled"`
	// ScannedAt is when the app's images were last looked for, and Error why
	// none could be scanned then. Nil before the first scan.
	ScannedAt *time.Time `json:"scannedAt,omitempty"`
	Error     string     `json:"error,omitempty"`
	// Summary counts the findings of every image, after ignores.
	Summary *SummaryResp `json:"summary"`
	Images  []*ImageResp `json:"images"`
	// Findings are those no ignore sets aside, the worst first; Ignored, the
	// rest, each with its ignore.
	Findings []*FindingResp        `json:"findings"`
	Ignored  []*IgnoredFindingResp `json:"ignored"`
	// Task is the app's scan not ended yet: queued or running.
	Task *ScanTaskResp `json:"task,omitempty"`
}

type ScanTaskResp struct {
	ID     string          `json:"id"`
	Status base.TaskStatus `json:"status"`
}
