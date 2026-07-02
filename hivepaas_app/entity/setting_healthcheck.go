package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentHealthcheckVersion = 1
)

var _ = registerSettingParser(base.SettingTypeHealthcheck, &healthcheckParser{})

type healthcheckParser struct {
}

func (s *healthcheckParser) New() SettingData {
	return &Healthcheck{}
}

type Healthcheck struct {
	HealthcheckType base.HealthcheckType     `json:"healthcheckType"`
	Interval        timeutil.Duration        `json:"interval"`
	MaxRetry        int                      `json:"maxRetry,omitempty"`
	RetryDelay      timeutil.Duration        `json:"retryDelay,omitempty"`
	Timeout         timeutil.Duration        `json:"timeout,omitempty"`
	SaveResultTasks bool                     `json:"saveResultTasks,omitempty"`
	REST            *HealthcheckREST         `json:"rest,omitempty"`
	GRPC            *HealthcheckGRPC         `json:"grpc,omitempty"`
	Notification    *HealthcheckNotification `json:"notification,omitempty"`
}

type HealthcheckREST struct {
	URL         string                     `json:"url"`
	Method      base.HTTPMethod            `json:"method,omitempty"`
	ContentType string                     `json:"contentType,omitempty"`
	Body        string                     `json:"body,omitempty"`
	ReturnCode  []int                      `json:"returnCode,omitempty"`
	ReturnText  *HealthcheckRESTReturnText `json:"returnText,omitempty"`
	ReturnJSON  *HealthcheckRESTReturnJSON `json:"returnJSON,omitempty"`
}

type HealthcheckRESTReturnText struct {
	Exact string `json:"exact,omitempty"`
	Regex string `json:"regex,omitempty"`
}

type HealthcheckRESTReturnJSON struct {
	Exact   any `json:"exact,omitempty"`
	Contain any `json:"contain,omitempty"`
}

type HealthcheckGRPC struct {
	Version      base.HealthcheckGRPCVersion `json:"version"`
	Addr         string                      `json:"addr"`
	Service      string                      `json:"service"`
	ReturnStatus base.HealthcheckGRPCStatus  `json:"returnStatus"`
}

type HealthcheckNotification struct {
	*BaseEventNotification
	MinSendInterval timeutil.Duration `json:"minSendInterval,omitempty"`
}

func (s *Healthcheck) GetType() base.SettingType {
	return base.SettingTypeHealthcheck
}

func (s *Healthcheck) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *Healthcheck) CalcResLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().CalcResLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Healthcheck) Migrate(setting *Setting) (hasChange bool, err error) {
	if setting.Version == CurrentHealthcheckVersion {
		return false, nil
	}
	if setting.Version > CurrentHealthcheckVersion {
		return false, apperrors.New(apperrors.ErrDataVerNewerThanSystemVer)
	}

	// TODO: add migration if we make any change

	setting.Version = CurrentHealthcheckVersion
	setting.UpdateVer++
	setting.MustSetData(s)
	return true, nil
}

func (s *Setting) AsHealthcheck() (*Healthcheck, error) {
	return parseSettingAs[*Healthcheck](s)
}

func (s *Setting) MustAsHealthcheck() *Healthcheck {
	return gofn.Must(s.AsHealthcheck())
}
