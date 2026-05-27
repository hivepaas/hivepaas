package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"
)

const (
	defaultDirMode = 0755
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

// Scheduled jobs

func (cfg *Config) DashboardGlobalSchedTaskDetailsURL(schedJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "settings", "sched-jobs", schedJobID,
		"tasks", taskID)) // TODO: update this
}

func (cfg *Config) DashboardAppSchedTaskDetailsURL(appID, projectID, schedJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "apps", appID,
		"sched-jobs", schedJobID, "tasks", taskID)) // TODO: update this
}

func (cfg *Config) DashboardProjectSchedTaskDetailsURL(projectID, schedJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "sched-jobs", schedJobID,
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

// Tasks

func (cfg *Config) DashboardTaskDetailsURL(taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "tasks", taskID)) // TODO: update this
}

/// BACK-END

func (cfg *Config) SsoBaseCallbackURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "auth/sso/callback"))
}

func (cfg *Config) SsoCallbackURL(id string) string {
	return gofn.Must(url.JoinPath(cfg.SsoBaseCallbackURL(), id))
}

func (cfg *Config) RepoWebhookURL(webhookID, secret string) string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "webhooks", webhookID, secret))
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

/// SSL CERTS

func (cfg *Config) DataPathSsl() string {
	return filepath.Join(cfg.AppPath, "ssl")
}
func (cfg *Config) DataPathSslCerts() string {
	return filepath.Join(cfg.DataPathSsl(), "certs")
}
func (cfg *Config) DataPathSslLetsEncrypt() string {
	return filepath.Join(cfg.DataPathSsl(), "letsencrypt")
}
func (cfg *Config) HttpPathSslLetsEncrypt() string {
	return "/letsencrypt/"
}

/// TRAEFIK

func (cfg *Config) DataPathTraefik() string {
	return filepath.Join(cfg.AppPath, "traefik")
}
func (cfg *Config) DataPathTraefikEtc() string {
	return filepath.Join(cfg.DataPathTraefik(), "etc")
}
func (cfg *Config) DataPathTraefikEtcDynamic() string {
	return filepath.Join(cfg.DataPathTraefikEtc(), "dynamic")
}

/// SYSTEM BACKUP

func (cfg *Config) DataPathSystemBackup() string {
	return filepath.Join(cfg.AppPath, "system", "backup")
}
func (cfg *Config) DataPathSystemBackupFiles() string {
	return filepath.Join(cfg.DataPathSystemBackup(), "files")
}

/// DIRS TO CREATE AT STARTUP

func (cfg *Config) DataPathsToInitAtStartup() map[string]os.FileMode {
	return map[string]os.FileMode{
		cfg.DataPathSslCerts():       defaultDirMode,
		cfg.DataPathSslLetsEncrypt(): defaultDirMode,

		cfg.DataPathTraefikEtcDynamic(): defaultDirMode,

		cfg.DataPathSystemBackupFiles(): defaultDirMode,
	}
}
