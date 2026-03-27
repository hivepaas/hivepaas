package config

import (
	"fmt"
	"net/url"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/base"
)

func (cfg *Config) BaseAPIURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, cfg.HTTPServer.BasePath))
}

/// FRONT-END DASHBOARD

// Users

func (cfg *Config) DashboardSsoSuccessURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "auth/sso/success"))
}

func (cfg *Config) DashboardUserSignupURL(token string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "auth/sign-up")) +
		fmt.Sprintf("?token=%s", token)
}

func (cfg *Config) DashboardPasswordResetURL(userID, token string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "auth/reset-password")) +
		fmt.Sprintf("?userID=%s&token=%s", userID, token)
}

// Deployments

func (cfg *Config) DashboardDeploymentDetailsURL(deploymentID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "deployments", deploymentID)) // TODO: update this
}

// Cron jobs

func (cfg *Config) DashboardGlobalCronTaskDetailsURL(cronJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "settings", "cron-jobs", cronJobID,
		"tasks", taskID)) // TODO: update this
}

func (cfg *Config) DashboardAppCronTaskDetailsURL(appID, projectID, cronJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "apps", appID,
		"cron-jobs", cronJobID, "tasks", taskID)) // TODO: update this
}

func (cfg *Config) DashboardProjectCronTaskDetailsURL(projectID, cronJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "cron-jobs", cronJobID,
		"tasks", taskID)) // TODO: update this
}

// Github Apps

func (cfg *Config) DashboardGlobalGithubAppsURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "settings/github-apps"))
}

func (cfg *Config) DashboardProjectGithubAppsURL(projectID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "github-apps"))
}

// Health checks

func (cfg *Config) DashboardAppHealthcheckDetailsURL(appID, projectID, healthcheckID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "apps", appID,
		"healthcheck", healthcheckID, "tasks", taskID)) // TODO: update this
}

func (cfg *Config) DashboardProjectHealthcheckDetailsURL(projectID, healthcheckID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "healthcheck", healthcheckID,
		"tasks", taskID)) // TODO: update this
}

/// BACK-END

func (cfg *Config) SsoBaseCallbackURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "auth/sso/callback"))
}

func (cfg *Config) SsoCallbackURL(id string) string {
	return gofn.Must(url.JoinPath(cfg.SsoBaseCallbackURL(), id))
}

func (cfg *Config) RepoWebhookURL(kind base.WebhookKind, secret string) string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "webhooks", string(kind), secret))
}

func (cfg *Config) GlobalGithubAppManifestFlowBeginURL(settingID, state string) string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "settings/github-apps", settingID,
		"manifest-flow/begin")) + "?state=" + state
}

func (cfg *Config) GlobalGithubAppManifestFlowProgressURL(settingID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "settings/github-apps", settingID,
		"manifest-flow/progress"))
}

func (cfg *Config) ProjectGithubAppManifestFlowBeginURL(projectID, settingID, state string) string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "projects", projectID,
		"github-apps", settingID, "manifest-flow/begin")) + "?state=" + state
}

func (cfg *Config) ProjectGithubAppManifestFlowProgressURL(projectID, settingID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "projects", projectID,
		"github-apps", settingID, "manifest-flow/progress"))
}

/// OBJECT PHOTOS

func (cfg *Config) DataPathPhoto() string {
	return filepath.Join(cfg.AppPath, "files", "photo")
}
func (cfg *Config) HttpPathPhoto() string {
	return "/files/photo/"
}

func (cfg *Config) DataPathUserPhoto() string {
	return filepath.Join(cfg.DataPathPhoto(), "user")
}
func (cfg *Config) DataPathProjectPhoto() string {
	return filepath.Join(cfg.DataPathPhoto(), "project")
}

/// SSL CERTS

func (cfg *Config) DataPathCerts() string {
	return filepath.Join(cfg.AppPath, "certs")
}

/// NGINX

func (cfg *Config) DataPathNginx() string {
	return filepath.Join(cfg.AppPath, "nginx")
}
func (cfg *Config) DataPathNginxEtc() string {
	return filepath.Join(cfg.DataPathNginx(), "etc")
}
func (cfg *Config) DataPathNginxEtcConf() string {
	return filepath.Join(cfg.DataPathNginxEtc(), "conf.d")
}
func (cfg *Config) DataPathNginxShare() string {
	return filepath.Join(cfg.DataPathNginx(), "share")
}
func (cfg *Config) DataPathNginxShareDomains() string {
	return filepath.Join(cfg.DataPathNginxShare(), "domains")
}
func (cfg *Config) DataPathNginxShareBasicAuth() string {
	return filepath.Join(cfg.DataPathNginxShare(), "basic-auth")
}

/// LETS ENCRYPT

func (cfg *Config) DataPathLetsEncrypt() string {
	return filepath.Join(cfg.AppPath, "letsencrypt")
}
func (cfg *Config) DataPathLetsEncryptEtc() string {
	return filepath.Join(cfg.DataPathLetsEncrypt(), "etc")
}

/// SYSTEM BACKUP

func (cfg *Config) DataPathSystemBackup() string {
	return filepath.Join(cfg.AppPath, "system", "backup")
}

func (cfg *Config) DataPathSystemBackupFiles() string {
	return filepath.Join(cfg.DataPathSystemBackup(), "files")
}
