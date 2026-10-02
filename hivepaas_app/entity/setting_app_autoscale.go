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

	// The bounds and defaults of a function's autoscale: see the spec,
	// docs/superpowers/specs/2026-10-03-function-autoscale-design.md.
	AppAutoscaleMinReplicasDefault  = 1
	AppAutoscaleMaxReplicasDefault  = 5
	AppAutoscaleMaxReplicasLimit    = 50
	AppAutoscaleTargetDefault       = 70
	AppAutoscaleTargetMin           = 10
	AppAutoscaleTargetMax           = 100
	AppAutoscaleScaleInDelayDefault = 5 * time.Minute
	AppAutoscaleScaleInDelayMin     = time.Minute
	AppAutoscaleScaleInDelayMax     = time.Hour
)

var _ = registerSettingParser(base.SettingTypeAppAutoscale, &appAutoscaleParser{})

type appAutoscaleParser struct{}

func (s *appAutoscaleParser) New() SettingData {
	return &AppAutoscale{}
}

// AppAutoscale is how an app's replicas follow its load, between Min and Max.
// Target is the share of an instance's Concurrency it is scaled to keep busy,
// in percent; ScaleInDelay how long the load stays low before it scales in.
type AppAutoscale struct {
	Enabled      bool              `json:"enabled"`
	MinReplicas  int               `json:"minReplicas"`
	MaxReplicas  int               `json:"maxReplicas"`
	Target       int               `json:"target"`
	ScaleInDelay timeutil.Duration `json:"scaleInDelay"`
}

// NewAppAutoscale is the settings as they start, off.
func NewAppAutoscale() *AppAutoscale {
	return &AppAutoscale{
		MinReplicas:  AppAutoscaleMinReplicasDefault,
		MaxReplicas:  AppAutoscaleMaxReplicasDefault,
		Target:       AppAutoscaleTargetDefault,
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
