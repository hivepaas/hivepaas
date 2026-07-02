package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
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

// App deployments

func (cfg *Config) DashboardAppDeploymentDetailsURL(appID, projectID, deploymentID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "apps", appID,
		"deployments", deploymentID))
}

// Scheduled jobs

func (cfg *Config) DashboardGlobalSchedTaskDetailsURL(schedJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "settings", "sched-jobs", schedJobID,
		"tasks", taskID)) // TODO: update this
}

func (cfg *Config) DashboardAppSchedTaskDetailsURL(projectID, appID, schedJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "apps", appID,
		"sched-jobs", schedJobID, "tasks", taskID))
}

func (cfg *Config) DashboardProjectSchedTaskDetailsURL(projectID, schedJobID, taskID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "sched-jobs", schedJobID,
		"tasks", taskID)) // TODO: update this
}

// Github Apps

func (cfg *Config) DashboardGlobalGithubAppsURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "sources/github-apps"))
}

func (cfg *Config) DashboardProjectGithubAppsURL(projectID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseURL, "projects", projectID, "provider-settings/github-apps"))
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

func (cfg *Config) RepoWebhookURL(webhookID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), "webhooks", webhookID))
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

/// LOCAL PATH

type LocalPath string

func (lp LocalPath) RelPath() string {
	return string(lp)
}
func (lp LocalPath) AbsPath() string {
	return filepath.Join(Current.AppPath, string(lp))
}
func (lp LocalPath) Join(elem ...string) LocalPath {
	return LocalPath(filepath.Join(append([]string{string(lp)}, elem...)...))
}

/// SSL CERTS

func (cfg *Config) DataPathSsl() LocalPath {
	return "ssl"
}
func (cfg *Config) DataPathSslCerts() LocalPath {
	return cfg.DataPathSsl().Join("certs")
}
func (cfg *Config) DataPathSslAcme() LocalPath {
	return cfg.DataPathSsl().Join("acme")
}
func (cfg *Config) HttpPathSslAcme() string {
	return "/acme/"
}

/// TRAEFIK

func (cfg *Config) DataPathTraefik() LocalPath {
	return "traefik"
}
func (cfg *Config) DataPathTraefikEtc() LocalPath {
	return cfg.DataPathTraefik().Join("etc")
}
func (cfg *Config) DataPathTraefikEtcDynamic() LocalPath {
	return cfg.DataPathTraefikEtc().Join("dynamic")
}

/// SYSTEM BACKUP

func (cfg *Config) DataPathSystemBackup() LocalPath {
	return LocalPath(filepath.Join("system", "backup"))
}
func (cfg *Config) DataPathSystemBackupFiles() LocalPath {
	return cfg.DataPathSystemBackup().Join("files")
}

/// SYSTEM CACHE

func (cfg *Config) DataPathSystemCache() LocalPath {
	return LocalPath(filepath.Join("system", "cache"))
}
func (cfg *Config) DataPathSystemCacheRepos() LocalPath {
	return cfg.DataPathSystemCache().Join("repos")
}

/// UPLOAD FILES

func (cfg *Config) DataPathFiles() LocalPath {
	return "files"
}

/// DIRS TO CREATE AT STARTUP

func (cfg *Config) DataPathsToInitAtStartup() map[string]os.FileMode {
	return map[string]os.FileMode{
		cfg.DataPathSslCerts().AbsPath(): base.DirModeDefault,
		cfg.DataPathSslAcme().AbsPath():  base.DirModeDefault,

		cfg.DataPathTraefikEtcDynamic().AbsPath(): base.DirModeDefault,

		cfg.DataPathSystemBackupFiles().AbsPath(): base.DirModeDefault,
		cfg.DataPathSystemCacheRepos().AbsPath():  base.DirModeDefault,
		cfg.DataPathFiles().AbsPath():             base.DirModeDefault,
	}
}
