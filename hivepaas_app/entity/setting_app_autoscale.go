package entity

import (
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentAppAutoscaleVersion = 1

	// The bounds and defaults of an app's autoscale: see the specs,
	// docs/superpowers/specs/2026-10-03-function-autoscale-design.md and
	// 2026-10-03-app-autoscale-design.md.
	AppAutoscaleMinReplicasDefault  = 1
	AppAutoscaleMaxReplicasDefault  = 5
	AppAutoscaleMaxReplicasLimit    = 50
	AppAutoscaleTargetDefault       = 70
	AppAutoscaleTargetMin           = 10
	AppAutoscaleTargetMax           = 100
	AppAutoscaleScaleInDelayDefault = 5 * time.Minute
	AppAutoscaleScaleInDelayMin     = time.Minute
	AppAutoscaleScaleInDelayMax     = time.Hour
	AppAutoscaleRequestsTargetMin   = 1
	AppAutoscaleRequestsTargetMax   = 1000
	AppAutoscaleCPUTargetDefault    = 70
)

var _ = registerSettingParser(base.SettingTypeAppAutoscale, &appAutoscaleParser{})

type appAutoscaleParser struct{}

func (s *appAutoscaleParser) New() SettingData {
	return &AppAutoscale{}
}

// AppAutoscale is how an app's replicas follow its load, between Min and Max;
// ScaleInDelay is how long the load stays low before it scales in.
//
// A function scales on its calls: Target is the share of an instance's
// Concurrency it is kept at, in percent. Any other app scales on its requests,
// its CPU or both: RequestsTarget is the requests in flight an instance takes,
// CPUTarget the share of an instance's CPU limit - or reservation - kept busy,
// in percent; 0 for a signal it does not scale on.
type AppAutoscale struct {
	Enabled        bool              `json:"enabled"`
	MinReplicas    int               `json:"minReplicas"`
	MaxReplicas    int               `json:"maxReplicas"`
	Target         int               `json:"target"`
	RequestsTarget int               `json:"requestsTarget,omitempty"`
	CPUTarget      int               `json:"cpuTarget,omitempty"`
	ScaleInDelay   timeutil.Duration `json:"scaleInDelay"`
}

// NewAppAutoscale is the settings as they start, off.
func NewAppAutoscale() *AppAutoscale {
	return &AppAutoscale{
		MinReplicas:  AppAutoscaleMinReplicasDefault,
		MaxReplicas:  AppAutoscaleMaxReplicasDefault,
		Target:       AppAutoscaleTargetDefault,
		CPUTarget:    AppAutoscaleCPUTargetDefault,
		ScaleInDelay: timeutil.Duration(AppAutoscaleScaleInDelayDefault),
	}
}

func (s *AppAutoscale) GetType() base.SettingType {
	return base.SettingTypeAppAutoscale
}

func (s *AppAutoscale) GetRefObjectIDs() *RefObjectIDs {
	return &RefObjectIDs{}
}

func (s *AppAutoscale) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *AppAutoscale) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentAppAutoscaleVersion {
		return false, nil
	}
	if setting.Version > CurrentAppAutoscaleVersion {
		return false, hperrors.Wrap(hperrors.ErrDataVerNewerThanSystemVer)
	}
	setting.Version = CurrentAppAutoscaleVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsAppAutoscale() (*AppAutoscale, error) {
	return parseSettingAs[*AppAutoscale](s)
}

func (s *Setting) MustAsAppAutoscale() *AppAutoscale {
	return gofn.Must(s.AsAppAutoscale())
}
