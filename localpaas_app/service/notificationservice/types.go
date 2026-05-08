package notificationservice

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

type TemplateType string

const (
	TemplateTypeEmail   TemplateType = "email"
	TemplateTypeSlack   TemplateType = "slack"
	TemplateTypeDiscord TemplateType = "discord"
)

type TemplateName string

const (
	TemplateAppDeploymentNotification TemplateName = "app-deployment-notification"
	TemplateCronTaskNotification      TemplateName = "cron-job-notification"
	TemplateHealthcheckNotification   TemplateName = "healthcheck-notification"
	TemplateSSLExpiringNotification   TemplateName = "ssl-expiring-notification"
	TemplateSSLRenewalNotification    TemplateName = "ssl-renewal-notification"
	TemplateSystemUpdateNotification  TemplateName = "system-update-notification"
)

type TemplateData interface {
	GetTitle() string
}

type BaseTemplateData struct {
	Title string
}

func (d *BaseTemplateData) GetTitle() string {
	return d.Title
}

type TaskResultNotificationReq struct {
	ActionSucceeded bool
	ScopeProject    *entity.Project
	ScopeApp        *entity.App
	ScopeUser       *entity.User
	RefObjects      *entity.RefObjects

	Notification *entity.Notification
	TemplateName TemplateName
	TemplateData TemplateData

	LastSendEvent     string // `success`, `failure`
	LastSendTimestamp time.Time
	LastEmailSent     bool
	LastSlackSent     bool
	LastDiscordSent   bool
}

type TaskResultNotificationResp struct {
	EmailSent   bool
	SlackSent   bool
	DiscordSent bool
}

//
// APP DEPLOYMENT
//

type TemplateDataAppDeployment struct {
	BaseTemplateData
	ProjectName   string
	AppName       string
	Succeeded     bool
	Method        base.DeploymentMethod
	RepoURL       string
	RepoRef       string
	CommitMsg     string
	Image         string
	StartedAt     time.Time
	Duration      time.Duration
	DashboardLink string
}

//
// CRON TASK
//

type TemplateDataCronTask struct {
	BaseTemplateData
	ProjectName   string
	AppName       string
	Succeeded     bool
	CronJobName   string
	Schedule      string
	CreatedAt     time.Time
	StartedAt     time.Time
	Duration      time.Duration
	Retries       int
	DashboardLink string
}

//
// HEALTH CHECK
//

type TemplateDataHealthcheck struct {
	BaseTemplateData
	ProjectName     string
	AppName         string
	Succeeded       bool
	HealthcheckName string
	HealthcheckType base.HealthcheckType
	StartedAt       time.Time
	Duration        time.Duration
	Retries         int
	Expect          string
	Actual          string
	DashboardLink   string
}

//
// SSL EXPIRING
//

type TemplateDataSSLExpiring struct {
	BaseTemplateData
	ProjectName   string
	AppName       string
	SSLName       string
	SSLType       string
	Domain        string
	CreatedAt     time.Time
	ExpireAt      time.Time
	ExpireIn      timeutil.Duration
	DashboardLink string
}

//
// SSL RENEWAL
//

type TemplateDataSSLRenewal struct {
	BaseTemplateData
	ProjectName   string
	AppName       string
	Succeeded     bool
	SSLName       string
	SSLType       string
	Domain        string
	CreatedAt     time.Time
	ExpireAt      time.Time
	NextRenewalIn timeutil.Duration
	DashboardLink string
}

//
// SYSTEM UPDATE
//

type TemplateDataSystemUpdate struct {
	BaseTemplateData
	Succeeded      bool
	CurrentVersion string
	TargetVersion  string
	StartedAt      time.Time
	Duration       time.Duration
	DashboardLink  string
}
