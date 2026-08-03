package entity

import (
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

const (
	CurrentPeriodicJobVersion = 1
)

var _ = registerSettingParser(base.SettingTypePeriodicJob, &periodicJobParser{})

type periodicJobParser struct {
}

func (s *periodicJobParser) New() SettingData {
	return &PeriodicJob{}
}

type PeriodicJob struct {
	Interval     timeutil.Duration     `json:"interval"`
	MaxRetry     int                   `json:"maxRetry,omitempty"`
	RetryDelay   timeutil.Duration     `json:"retryDelay,omitempty"`
	Timeout      timeutil.Duration     `json:"timeout,omitempty"`
	Healthcheck  *PeriodicHealthcheck  `json:"healthcheck,omitempty"`
	Notification *PeriodicNotification `json:"notification,omitempty"`
}

type PeriodicHealthcheck struct {
	HealthcheckType base.HealthcheckType `json:"healthcheckType"`
	REST            *HealthcheckREST     `json:"rest,omitempty"`
	GRPC            *HealthcheckGRPC     `json:"grpc,omitempty"`
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

type PeriodicNotification struct {
	*BaseEventNotification
	MinSendInterval timeutil.Duration `json:"minSendInterval,omitempty"`
}

func (s *PeriodicJob) GetType() base.SettingType {
	return base.SettingTypePeriodicJob
}

func (s *PeriodicJob) GetRefObjectIDs() *RefObjectIDs {
	refIDs := &RefObjectIDs{}
	if s.Notification != nil {
		refIDs.AddRefIDs(s.Notification.GetRefObjectIDs())
	}
	return refIDs
}

func (s *PeriodicJob) GetResourceLinks(setting *Setting) []*ResLink {
	return s.GetRefObjectIDs().GetResourceLinks(base.ResourceTypeSetting, setting.ID)
}

func (s *Setting) AsPeriodicJob() (*PeriodicJob, error) {
	return parseSettingAs[*PeriodicJob](s)
}

func (s *Setting) MustAsPeriodicJob() *PeriodicJob {
	return gofn.Must(s.AsPeriodicJob())
}
