package notificationservice

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/infra/database"
)

type Service interface {
	// App deployment notification
	EmailSendAppDeploymentNotification(ctx context.Context, db database.IDB,
		data *EmailMsgDataAppDeploymentNotification) error
	SlackSendAppDeploymentNotification(ctx context.Context, db database.IDB,
		data *SlackMsgDataAppDeploymentNotification) error
	DiscordSendAppDeploymentNotification(ctx context.Context, db database.IDB,
		data *DiscordMsgDataAppDeploymentNotification) error

	// Cron job notification
	EmailSendCronTaskNotification(ctx context.Context, db database.IDB,
		data *EmailMsgDataCronTaskNotification) error
	SlackSendCronTaskNotification(ctx context.Context, db database.IDB,
		data *SlackMsgDataCronTaskNotification) error
	DiscordSendCronTaskNotification(ctx context.Context, db database.IDB,
		data *DiscordMsgDataCronTaskNotification) error

	// SSL renewal
	EmailSendSSLExpiringNotification(ctx context.Context, db database.IDB,
		data *EmailMsgDataSSLExpiringNotification) error
	SlackSendSSLExpiringNotification(ctx context.Context, db database.IDB,
		data *SlackMsgDataSSLExpiringNotification) error
	DiscordSendSSLExpiringNotification(ctx context.Context, db database.IDB,
		data *DiscordMsgDataSSLExpiringNotification) error
	EmailSendSSLRenewalNotification(ctx context.Context, db database.IDB,
		data *EmailMsgDataSSLRenewalNotification) error
	SlackSendSSLRenewalNotification(ctx context.Context, db database.IDB,
		data *SlackMsgDataSSLRenewalNotification) error
	DiscordSendSSLRenewalNotification(ctx context.Context, db database.IDB,
		data *DiscordMsgDataSSLRenewalNotification) error

	// Healthcheck
	EmailSendHealthcheckNotification(ctx context.Context, db database.IDB,
		data *EmailMsgDataHealthcheckNotification) error
	SlackSendHealthcheckNotification(ctx context.Context, db database.IDB,
		data *SlackMsgDataHealthcheckNotification) error
	DiscordSendHealthcheckNotification(ctx context.Context, db database.IDB,
		data *DiscordMsgDataHealthcheckNotification) error

	// System update notification
	EmailSendSystemUpdateNotification(ctx context.Context, db database.IDB,
		data *EmailMsgDataSystemUpdateNotification) error
	SlackSendSystemUpdateNotification(ctx context.Context, db database.IDB,
		data *SlackMsgDataSystemUpdateNotification) error
	DiscordSendSystemUpdateNotification(ctx context.Context, db database.IDB,
		data *DiscordMsgDataSystemUpdateNotification) error
}
