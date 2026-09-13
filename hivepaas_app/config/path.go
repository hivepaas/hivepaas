package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

const (
	beBasePathSettings = "settings"
)

func (cfg *Config) BaseAPIURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseURL(), cfg.HTTPServer.BasePath))
}

func (cfg *Config) BaseDashboardURL() string {
	return cfg.BaseURL()
}

/// FRONT-END DASHBOARD

// Users

func (cfg *Config) DashboardSsoSuccessURL() string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "auth/sso/success"))
}

func (cfg *Config) DashboardUserSignupURL(token string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "auth/sign-up")) +
		fmt.Sprintf("?token=%s", token)
}

func (cfg *Config) DashboardPasswordResetURL(userID, token string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), "auth/reset-password")) +
		fmt.Sprintf("?userID=%s&token=%s", userID, token)
}

// App deployments

func (cfg *Config) DashboardAppDeploymentDetailsURL(basePath, deploymentID string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), basePath, "deployments", deploymentID))
}

// Github Apps

func (cfg *Config) DashboardGithubAppsURL(basePath string) string {
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), basePath, "integrations/github-apps"))
}

// Tasks

func (cfg *Config) DashboardTaskDetailsURL(basePath, taskID string, scope base.ObjectScopeType) string {
	switch scope {
	case base.ObjectScopeProject, base.ObjectScopeProjectEnv, base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		basePath += "/operations"
	case base.ObjectScopeApp, base.ObjectScopeUser:
	}
	return gofn.Must(url.JoinPath(cfg.BaseDashboardURL(), basePath, "tasks", taskID))
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

func (cfg *Config) GithubAppManifestFlowBeginURL(basePath, settingID, state string) string {
	if basePath == "" {
		basePath = beBasePathSettings // global scope
	}
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), basePath, "github-apps", settingID,
		"manifest-flow/begin")) + "?state=" + state
}

func (cfg *Config) GithubAppManifestFlowProgressURL(basePath, settingID string) string {
	if basePath == "" {
		basePath = beBasePathSettings // global scope
	}
	return gofn.Must(url.JoinPath(cfg.BaseAPIURL(), basePath, "github-apps", settingID,
		"manifest-flow/progress"))
}

/// LOCAL PATH

type LocalPath string

func (lp LocalPath) RelPath() string {
	return string(lp)
}
func (lp LocalPath) AbsPath() string {
	return filepath.Join(Current().AppPath, string(lp))
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

/// STATIC ICONS

func (cfg *Config) HttpPathStaticIcons() string {
	return "/static/icons/"
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
