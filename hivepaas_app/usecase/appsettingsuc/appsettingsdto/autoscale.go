package appsettingsdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionautoscaleservice"
)

type GetAppAutoscaleReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`
}

func NewGetAppAutoscaleReq() *GetAppAutoscaleReq {
	return &GetAppAutoscaleReq{}
}

func (req *GetAppAutoscaleReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 3) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppAutoscaleResp struct {
	Meta *basedto.Meta     `json:"meta"`
	Data *AppAutoscaleResp `json:"data"`
}

// AppAutoscaleResp is a function's autoscale: its settings, its replicas now,
// and - when it is on but cannot act - why it is paused.
type AppAutoscaleResp struct {
	Enabled      bool              `json:"enabled"`
	MinReplicas  int               `json:"minReplicas"`
	MaxReplicas  int               `json:"maxReplicas"`
	Target       int               `json:"target"`
	ScaleInDelay timeutil.Duration `json:"scaleInDelay"`
	// Replicas is the function's now; 0 when it is stopped.
	Replicas int `json:"replicas"`
	// Paused is why autoscale cannot act: its calls cannot be read (a reason
	// of the logs' history, such as disabled), or the function does not run a
	// set number of instances (not-replicated).
	Paused string `json:"paused,omitempty"`
	// Events are the latest scalings, the latest first.
	Events    []*AppAutoscaleEventResp `json:"events"`
	UpdateVer int                      `json:"updateVer"`
}

// AppAutoscaleEventResp is one scaling: when, from and to how many replicas,
// and what it was decided from - the calls in flight over the minute before,
// the calls, those turned away.
type AppAutoscaleEventResp struct {
	Time      time.Time `json:"time"`
	From      int       `json:"from"`
	To        int       `json:"to"`
	InFlight  float64   `json:"inFlight"`
	Calls     int64     `json:"calls"`
	Throttled int64     `json:"throttled"`
	Reason    string    `json:"reason"`
}

type UpdateAppAutoscaleReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	Enabled      bool              `json:"enabled"`
	MinReplicas  int               `json:"minReplicas"`
	MaxReplicas  int               `json:"maxReplicas"`
	Target       int               `json:"target"`
	ScaleInDelay timeutil.Duration `json:"scaleInDelay"`

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppAutoscaleReq() *UpdateAppAutoscaleReq {
	return &UpdateAppAutoscaleReq{}
}

func (req *UpdateAppAutoscaleReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 8) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateNumber(&req.MinReplicas, true, 1,
		entity.AppAutoscaleMaxReplicasLimit, "minReplicas")...)
	validators = append(validators, basedto.ValidateNumber(&req.MaxReplicas, true, max(req.MinReplicas, 1),
		entity.AppAutoscaleMaxReplicasLimit, "maxReplicas")...)
	validators = append(validators, basedto.ValidateNumber(&req.Target, true, entity.AppAutoscaleTargetMin,
		entity.AppAutoscaleTargetMax, "target")...)
	validators = append(validators, basedto.ValidateDuration(&req.ScaleInDelay, true,
		timeutil.Duration(entity.AppAutoscaleScaleInDelayMin), timeutil.Duration(entity.AppAutoscaleScaleInDelayMax),
		"scaleInDelay")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *UpdateAppAutoscaleReq) ToEntity() *entity.AppAutoscale {
	return &entity.AppAutoscale{
		Enabled: req.Enabled, MinReplicas: req.MinReplicas, MaxReplicas: req.MaxReplicas, Target: req.Target,
		ScaleInDelay: req.ScaleInDelay,
	}
}

type UpdateAppAutoscaleResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// TransformAppAutoscale answers the settings, the defaults when there are none.
func TransformAppAutoscale(
	setting *entity.Setting, replicas int, paused string, events []*functionautoscaleservice.Event,
) (*AppAutoscaleResp, error) {
	autoscale := entity.NewAppAutoscale()
	resp := &AppAutoscaleResp{Replicas: replicas, Paused: paused,
		Events: make([]*AppAutoscaleEventResp, 0, len(events))}
	for _, e := range events {
		resp.Events = append(resp.Events, &AppAutoscaleEventResp{Time: e.Time, From: e.From, To: e.To,
			InFlight: e.InFlight, Calls: e.Calls, Throttled: e.Throttled, Reason: e.Reason})
	}
	if setting != nil {
		var err error
		if autoscale, err = setting.AsAppAutoscale(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		resp.UpdateVer = setting.UpdateVer
	}
	resp.Enabled, resp.MinReplicas, resp.MaxReplicas = autoscale.Enabled, autoscale.MinReplicas, autoscale.MaxReplicas
	resp.Target, resp.ScaleInDelay = autoscale.Target, autoscale.ScaleInDelay
	if !resp.Enabled {
		resp.Paused = ""
	}
	return resp, nil
}

// ModifyRequest gives a request that leaves the delay out the default one.
func (req *UpdateAppAutoscaleReq) ModifyRequest() error {
	if req.ScaleInDelay == 0 {
		req.ScaleInDelay = timeutil.Duration(entity.AppAutoscaleScaleInDelayDefault)
	}
	return nil
}
