package hpappdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

// GetHpAppUpdatePlanReq asks what an update to a published version would do.
type GetHpAppUpdatePlanReq struct {
	TargetVersion string `json:"-" mapstructure:"targetVersion"`
}

func NewGetHpAppUpdatePlanReq() *GetHpAppUpdatePlanReq {
	return &GetHpAppUpdatePlanReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetHpAppUpdatePlanReq) Validate() hperrors.ValidationErrors {
	validators := basedto.ValidateStr(&req.TargetVersion, true, 1, 100, "targetVersion") //nolint:mnd
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetHpAppUpdatePlanResp struct {
	Meta *basedto.Meta            `json:"meta"`
	Data *HpAppUpdatePlanDataResp `json:"data"`
}

type HpAppUpdatePlanDataResp struct {
	Current *hpappservice.CurrentRelease `json:"current"`
	Target  *UpdateTargetResp            `json:"target"`
	// Components are in the order the update reaches them.
	Components []*UpdateComponentResp `json:"components"`
	// Blocked is whether the update would refuse to run: a component would
	// cross a major version the release does not allow.
	Blocked bool `json:"blocked"`
	// RequiresBackup is whether the database backup cannot be skipped.
	RequiresBackup bool `json:"requiresBackup"`
}

type UpdateTargetResp struct {
	AppVersion  string        `json:"appVersion"`
	Channel     string        `json:"channel"`
	ReleaseDate timeutil.Date `json:"releaseDate"`
	NotesURL    string        `json:"notesUrl,omitempty"`
}

type UpdateComponentResp struct {
	// Key is db, redis, traefik, victoria-logs, vlagent, registry, app or worker.
	Key          string `json:"key"`
	CurrentImage string `json:"currentImage"`
	TargetImage  string `json:"targetImage"`
	// Change is none, update, major, blocked or not-deployed.
	Change string `json:"change"`
	// Reason is the update's own words for what it would do.
	Reason            string `json:"reason"`
	RequiresBackup    bool   `json:"requiresBackup"`
	InterruptsTraffic bool   `json:"interruptsTraffic"`
}

func TransformUpdatePlan(
	current *hpappservice.CurrentRelease,
	target *UpdateTargetResp,
	plan *sysupdateservice.UpdatePlan,
) *HpAppUpdatePlanDataResp {
	resp := &HpAppUpdatePlanDataResp{
		Current:        current,
		Target:         target,
		Components:     make([]*UpdateComponentResp, 0, len(plan.Components)),
		Blocked:        plan.Blocked,
		RequiresBackup: plan.RequiresBackup,
	}
	for _, c := range plan.Components {
		resp.Components = append(resp.Components, &UpdateComponentResp{
			Key:               c.Key,
			CurrentImage:      c.CurrentImage,
			TargetImage:       c.TargetImage,
			Change:            string(c.Change),
			Reason:            c.Reason,
			RequiresBackup:    c.RequiresBackup,
			InterruptsTraffic: c.InterruptsTraffic,
		})
	}
	return resp
}
