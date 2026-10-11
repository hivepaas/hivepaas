package notificationservice

import (
	"fmt"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

type TemplateType string

const (
	TemplateTypeEmail    TemplateType = "email"
	TemplateTypeSlack    TemplateType = "slack"
	TemplateTypeDiscord  TemplateType = "discord"
	TemplateTypeTelegram TemplateType = "telegram"
	TemplateTypeLark     TemplateType = "lark"
)

type TemplateName string

const (
	TemplateAppDeploymentNotification TemplateName = "app-deployment-notification"
	TemplateAppCloneNotification      TemplateName = "app-clone-notification"
	TemplateSchedTaskNotification     TemplateName = "sched-job-notification"
	TemplateHealthcheckNotification   TemplateName = "healthcheck-notification"
	TemplateSSLExpiringNotification   TemplateName = "ssl-expiring-notification"
	TemplateSSLRenewalNotification    TemplateName = "ssl-renewal-notification"
	TemplateSystemUpdateNotification  TemplateName = "system-update-notification"
	TemplateImageScanNotification     TemplateName = "image-scan-notification"
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
	Scope           *entity.ObjectScope
	RefObjects      *entity.RefObjects

	Notification *entity.Notification
	TemplateName TemplateName
	TemplateData TemplateData

	LastEvent  string // `success`, `failure`
	LastSendTs time.Time
}

type TaskResultNotificationResp struct {
	SendTs      time.Time
	DeliveryMap map[string]bool
}

func (r *TaskResultNotificationResp) HasSend() bool { // true if has at least one sending
	for _, sent := range r.DeliveryMap {
		if sent {
			return true
		}
	}
	return false
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
	CommitAuthor  string
	Image         string
	StartedAt     time.Time
	Duration      time.Duration
	DashboardLink string
	// Reason is why a failed deployment failed, as its details say it; empty
	// for one that succeeded.
	Reason string
}

func (d TemplateDataAppDeployment) StartedAtFormatted() string {
	return formatDateTime(d.StartedAt)
}

//
// APP CLONE
//

type TemplateDataAppClone struct {
	BaseTemplateData
	ProjectName string
	// AppName is the app cloned; CloneName and CloneEnv, its copy's.
	AppName       string
	CloneName     string
	CloneEnv      string
	Succeeded     bool
	StartedAt     time.Time
	Duration      time.Duration
	DashboardLink string
	// Reason is why a clone failed; empty for one that succeeded.
	Reason string
}

func (d TemplateDataAppClone) StartedAtFormatted() string {
	return formatDateTime(d.StartedAt)
}

//
// CRON TASK
//

type TemplateDataSchedTask struct {
	BaseTemplateData
	ProjectName   string
	AppName       string
	Succeeded     bool
	SchedJobName  string
	Schedule      string
	StartedAt     time.Time
	Duration      time.Duration
	Retries       int
	LastError     string
	DashboardLink string
}

func (d TemplateDataSchedTask) StartedAtFormatted() string {
	return formatDateTime(d.StartedAt)
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

func (d TemplateDataHealthcheck) StartedAtFormatted() string {
	return formatDateTime(d.StartedAt)
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

func (d TemplateDataSSLExpiring) CreatedAtFormatted() string {
	return formatDateTime(d.CreatedAt)
}

func (d TemplateDataSSLExpiring) ExpireAtFormatted() string {
	return formatDateTime(d.ExpireAt)
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

func (d TemplateDataSSLRenewal) CreatedAtFormatted() string {
	return formatDateTime(d.CreatedAt)
}

func (d TemplateDataSSLRenewal) ExpireAtFormatted() string {
	return formatDateTime(d.ExpireAt)
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

func (d TemplateDataSystemUpdate) StartedAtFormatted() string {
	return formatDateTime(d.StartedAt)
}

//
// IMAGE SCAN
//

// TemplateDataImageScan is the daily scan's report: the apps that run fixable
// CRITICAL or HIGH vulnerabilities their last report did not have.
type TemplateDataImageScan struct {
	BaseTemplateData
	Apps          []*ImageScanApp
	DashboardLink string
}

// ImageScanApp is an app of the report, and what is new in it.
type ImageScanApp struct {
	ProjectName string
	EnvName     string
	AppName     string
	// Critical and High are how many new fixable findings of each.
	Critical int
	High     int
	Link     string
}

// Where is the app as its screens name it: project / env / app.
func (a ImageScanApp) Where() string {
	return a.ProjectName + " / " + a.EnvName + " / " + a.AppName
}

// Counts is what is new: "2 CRITICAL, 1 HIGH".
func (a ImageScanApp) Counts() string {
	var parts []string
	if a.Critical > 0 {
		parts = append(parts, fmt.Sprintf("%d CRITICAL", a.Critical))
	}
	if a.High > 0 {
		parts = append(parts, fmt.Sprintf("%d HIGH", a.High))
	}
	return strings.Join(parts, ", ")
}

// Lines is the apps as text, one a line, for a channel that takes no table.
func (d TemplateDataImageScan) Lines() string {
	lines := make([]string, 0, len(d.Apps))
	for _, app := range d.Apps {
		lines = append(lines, "• "+app.Where()+": "+app.Counts())
	}
	return strings.Join(lines, "\n")
}
