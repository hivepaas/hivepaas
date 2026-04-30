package notificationservice

import (
	"time"

	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
)

//
// APP DEPLOYMENT
//

type BaseMsgDataAppDeploymentNotification struct {
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

type EmailMsgDataAppDeploymentNotification struct {
	*BaseMsgDataAppDeploymentNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

type SlackMsgDataAppDeploymentNotification struct {
	*BaseMsgDataAppDeploymentNotification
	Setting *entity.Slack
}

type DiscordMsgDataAppDeploymentNotification struct {
	*BaseMsgDataAppDeploymentNotification
	Setting *entity.Discord
}

//
// CRON TASK
//

type BaseMsgDataCronTaskNotification struct {
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

type EmailMsgDataCronTaskNotification struct {
	*BaseMsgDataCronTaskNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

type SlackMsgDataCronTaskNotification struct {
	*BaseMsgDataCronTaskNotification
	Setting *entity.Slack
}

type DiscordMsgDataCronTaskNotification struct {
	*BaseMsgDataCronTaskNotification
	Setting *entity.Discord
}

//
// HEALTH CHECK
//

type BaseMsgDataHealthcheckNotification struct {
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

type EmailMsgDataHealthcheckNotification struct {
	*BaseMsgDataHealthcheckNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

type SlackMsgDataHealthcheckNotification struct {
	*BaseMsgDataHealthcheckNotification
	Setting *entity.Slack
}

type DiscordMsgDataHealthcheckNotification struct {
	*BaseMsgDataHealthcheckNotification
	Setting *entity.Discord
}

//
// SSL EXPIRING
//

type BaseMsgDataSSLExpiringNotification struct {
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

type EmailMsgDataSSLExpiringNotification struct {
	*BaseMsgDataSSLExpiringNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

type SlackMsgDataSSLExpiringNotification struct {
	*BaseMsgDataSSLExpiringNotification
	Setting *entity.Slack
}

type DiscordMsgDataSSLExpiringNotification struct {
	*BaseMsgDataSSLExpiringNotification
	Setting *entity.Discord
}

//
// SSL RENEWAL
//

type BaseMsgDataSSLRenewalNotification struct {
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

type EmailMsgDataSSLRenewalNotification struct {
	*BaseMsgDataSSLRenewalNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

type SlackMsgDataSSLRenewalNotification struct {
	*BaseMsgDataSSLRenewalNotification
	Setting *entity.Slack
}

type DiscordMsgDataSSLRenewalNotification struct {
	*BaseMsgDataSSLRenewalNotification
	Setting *entity.Discord
}

//
// SYSTEM UPDATE
//

type BaseMsgDataSystemUpdateNotification struct {
	Succeeded      bool
	CurrentVersion string
	TargetVersion  string
	StartedAt      time.Time
	Duration       time.Duration
	DashboardLink  string
}

type EmailMsgDataSystemUpdateNotification struct {
	*BaseMsgDataSystemUpdateNotification
	Email      *entity.Email
	Recipients []string
	Subject    string
}

type SlackMsgDataSystemUpdateNotification struct {
	*BaseMsgDataSystemUpdateNotification
	Setting *entity.Slack
}

type DiscordMsgDataSystemUpdateNotification struct {
	*BaseMsgDataSystemUpdateNotification
	Setting *entity.Discord
}
