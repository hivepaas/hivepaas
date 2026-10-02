package appsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
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
	// of the logs' history, such as disabled).
	Paused    string `json:"paused,omitempty"`
	UpdateVer int    `json:"updateVer"`
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
func TransformAppAutoscale(setting *entity.Setting, replicas int, paused string) (*AppAutoscaleResp, error) {
	autoscale := entity.NewAppAutoscale()
	resp := &AppAutoscaleResp{Replicas: replicas, Paused: paused}
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
