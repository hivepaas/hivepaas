package appsettingsdto

import (
	"time"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appautoscaleservice"
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

// AppAutoscaleResp is an app's autoscale: its settings, its replicas now, why
// it cannot act, and its latest scalings.
//
// A function scales on its calls, at Target of its Concurrency; any other app
// on its requests, at RequestsTarget an instance, and its CPU, at CPUTarget of
// its limit - 0 for a signal it does not scale on.
type AppAutoscaleResp struct {
	IsFunction     bool              `json:"isFunction"`
	Enabled        bool              `json:"enabled"`
	MinReplicas    int               `json:"minReplicas"`
	MaxReplicas    int               `json:"maxReplicas"`
	Target         int               `json:"target"`
	RequestsTarget int               `json:"requestsTarget"`
	CPUTarget      int               `json:"cpuTarget"`
	ScaleInDelay   timeutil.Duration `json:"scaleInDelay"`
	// Replicas is the app's now; 0 when it is stopped.
	Replicas int `json:"replicas"`
	// Paused is why autoscale cannot act, on or off: a function's calls, or
	// all the signals an app scales on, cannot be read (a reason of the logs'
	// history, such as disabled, or of the signal); the app does not run a set
	// number of instances (not-replicated), or publishes a port on its node
	// (host-ports).
	Paused string `json:"paused,omitempty"`
	// RequestsUnavailable and CPUUnavailable are why an app's signal cannot
	// be read now, whether it scales on it or not: not-exposed, an access
	// log's reason, agent-unlabelled, identity-missing, no-cpu-limit, or a
	// reason of the logs' history. Empty when it can, and for a function.
	RequestsUnavailable string `json:"requestsUnavailable,omitempty"`
	CPUUnavailable      string `json:"cpuUnavailable,omitempty"`
	// Pending is how many of its tasks are wanted and not running: the
	// cluster may have no room for them, and autoscale holds its scale-outs.
	Pending int `json:"pending"`
	// WritableMounts is whether its replicas would share, or each have their
	// own, a volume or a bind mount it writes to.
	WritableMounts bool `json:"writableMounts"`
	// Events are the latest scalings, the latest first.
	Events    []*AppAutoscaleEventResp `json:"events"`
	UpdateVer int                      `json:"updateVer"`
}

// AppAutoscaleEventResp is one scaling: when, from and to how many replicas,
// and what it was decided from - the calls or requests in flight over the
// minute before; a function's calls and those turned away; an app's requests
// and CPU, in percent of its limit.
type AppAutoscaleEventResp struct {
	Time      time.Time `json:"time"`
	From      int       `json:"from"`
	To        int       `json:"to"`
	InFlight  float64   `json:"inFlight"`
	Calls     int64     `json:"calls"`
	Throttled int64     `json:"throttled"`
	Requests  int64     `json:"requests"`
	CPU       float64   `json:"cpu"`
	Reason    string    `json:"reason"`
}

type UpdateAppAutoscaleReq struct {
	ProjectID    string `json:"-"`
	ProjectEnvID string `json:"-"`
	AppID        string `json:"-"`

	Enabled        bool              `json:"enabled"`
	MinReplicas    int               `json:"minReplicas"`
	MaxReplicas    int               `json:"maxReplicas"`
	Target         int               `json:"target"`
	RequestsTarget int               `json:"requestsTarget"`
	CPUTarget      int               `json:"cpuTarget"`
	ScaleInDelay   timeutil.Duration `json:"scaleInDelay"`

	UpdateVer int `json:"updateVer"`
}

func NewUpdateAppAutoscaleReq() *UpdateAppAutoscaleReq {
	return &UpdateAppAutoscaleReq{}
}

func (req *UpdateAppAutoscaleReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, basedto.ValidateID(&req.ProjectID, true, "projectId")...)
	validators = append(validators, basedto.ValidateID(&req.ProjectEnvID, true, "projectEnv")...)
	validators = append(validators, basedto.ValidateID(&req.AppID, true, "appId")...)
	validators = append(validators, basedto.ValidateNumber(&req.MinReplicas, true, 1,
		entity.AppAutoscaleMaxReplicasLimit, "minReplicas")...)
	validators = append(validators, basedto.ValidateNumber(&req.MaxReplicas, true, max(req.MinReplicas, 1),
		entity.AppAutoscaleMaxReplicasLimit, "maxReplicas")...)
	validators = append(validators, basedto.ValidateNumber(&req.Target, true, entity.AppAutoscaleTargetMin,
		entity.AppAutoscaleTargetMax, "target")...)
	// 0 for a signal the app does not scale on: whether it scales on one at
	// least is the use case's to say, which knows whether it is a function.
	validators = append(validators, basedto.ValidateNumber(&req.RequestsTarget, false,
		entity.AppAutoscaleRequestsTargetMin, entity.AppAutoscaleRequestsTargetMax, "requestsTarget")...)
	validators = append(validators, basedto.ValidateNumber(&req.CPUTarget, false, entity.AppAutoscaleTargetMin,
		entity.AppAutoscaleTargetMax, "cpuTarget")...)
	validators = append(validators, basedto.ValidateDuration(&req.ScaleInDelay, true,
		timeutil.Duration(entity.AppAutoscaleScaleInDelayMin), timeutil.Duration(entity.AppAutoscaleScaleInDelayMax),
		"scaleInDelay")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

func (req *UpdateAppAutoscaleReq) ToEntity() *entity.AppAutoscale {
	return &entity.AppAutoscale{
		Enabled: req.Enabled, MinReplicas: req.MinReplicas, MaxReplicas: req.MaxReplicas, Target: req.Target,
		RequestsTarget: req.RequestsTarget, CPUTarget: req.CPUTarget, ScaleInDelay: req.ScaleInDelay,
	}
}

type UpdateAppAutoscaleResp struct {
	Meta *basedto.Meta `json:"meta"`
}

// AppAutoscaleState is what an app's autoscale answers besides its settings:
// its replicas now, whether it is a function, why it cannot act, and its
// latest scalings. Check is nil for a function.
type AppAutoscaleState struct {
	Replicas   int
	IsFunction bool
	Paused     string
	Check      *appautoscaleservice.Check
	Events     []*appautoscaleservice.Event
}

// TransformAppAutoscale answers the settings, the defaults when there are none.
func TransformAppAutoscale(setting *entity.Setting, st *AppAutoscaleState) (*AppAutoscaleResp, error) {
	autoscale := entity.NewAppAutoscale()
	resp := &AppAutoscaleResp{Replicas: st.Replicas, IsFunction: st.IsFunction, Paused: st.Paused,
		Events: make([]*AppAutoscaleEventResp, 0, len(st.Events))}
	if c := st.Check; c != nil {
		resp.RequestsUnavailable, resp.CPUUnavailable = c.Requests, c.CPU
		resp.Pending, resp.WritableMounts = c.Pending, c.WritableMounts
	}
	for _, e := range st.Events {
		resp.Events = append(resp.Events, &AppAutoscaleEventResp{Time: e.Time, From: e.From, To: e.To,
			InFlight: e.InFlight, Calls: e.Calls, Throttled: e.Throttled, Requests: e.Requests, CPU: e.CPU,
			Reason: e.Reason})
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
	resp.RequestsTarget, resp.CPUTarget = autoscale.RequestsTarget, autoscale.CPUTarget
	return resp, nil
}

// ModifyRequest gives a request that leaves the delay out the default one.
func (req *UpdateAppAutoscaleReq) ModifyRequest() error {
	if req.ScaleInDelay == 0 {
		req.ScaleInDelay = timeutil.Duration(entity.AppAutoscaleScaleInDelayDefault)
	}
	return nil
}
