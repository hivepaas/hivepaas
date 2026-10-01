package mcp

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/auditloguc/auditlogdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/networkuc/networkdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backupsnapshotuc/backupsnapshotdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/configfileuc/configfiledto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/gitcredentialuc/gitcredentialdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/notificationuc/notificationdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/registryauthuc/registryauthdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/secretuc/secretdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

// The settings a project keeps for its apps, and each env for its own: read
// where they are kept - a project's, or one env's, which also answers the ones
// it inherits from its project, marked inherited. Secrets are masked, as the
// dashboard shows them until somebody reveals them; nothing here reveals.

// atProjectOrEnv is the path of a list kept both by a project and by each env.
func atProjectOrEnv(path string) map[under]string {
	return map[under]string{underProject: path, underEnv: path}
}

const descProjectOrEnv = "Given an env, the env's own and those it inherits from its project; given only " +
	"the project, the project's."

// settingListParams are the parameters every list of settings takes, less its
// kind, which only some of them use.
func settingListParams(extra map[string]string) map[string]string {
	return pagingParams(mergeDescs(map[string]string{
		paramSearch: descSearchName,
		paramStatus: "only those in these states: " + settingStatuses,
	}, extra))
}

func mergeDescs(a, b map[string]string) map[string]string {
	for k, v := range b {
		a[k] = v
	}
	return a
}

func projectEndpoints() []getEndpoint { //nolint:funlen // a table
	return []getEndpoint{
		{
			name: "list_ssl_certs", title: "List certificates",
			description: "GET /projects/{project}[/{env}]/ssl-certs. The certificates domains are served " +
				"with: each one's domain, type (Let's Encrypt, custom, self-signed...), status, when it " +
				"expires and whether it renews; its certificate and key masked. Where a domain has no " +
				"HTTPS yet, this says whether a certificate covers it. " + descProjectOrEnv,
			paths: atProjectOrEnv("/ssl-certs"),
			query: &sslcertdto.ListSSLCertReq{},
			params: settingListParams(map[string]string{
				argKind:     "only certificates of these types: " + statusValues(base.AllSSLCertTypes),
				paramDomain: "only the certificates covering this domain",
			}),
			answer: func() any { return &sslcertdto.ListSSLCertResp{} },
		},
		{
			name: "list_config_files", title: "List config files",
			description: "GET /projects/{project}[/{env}]/config-files. The config files apps mount, each " +
				"with its content. " + descProjectOrEnv,
			paths:  atProjectOrEnv("/config-files"),
			query:  &configfiledto.ListConfigFileReq{},
			omit:   []string{argKind},
			params: settingListParams(nil),
			answer: func() any { return &configfiledto.ListConfigFileResp{} },
		},
		{
			name: "list_secrets", title: "List secrets",
			description: "GET /projects/{project}[/{env}]/secrets. The secrets apps use, by name; each " +
				"value masked. " + descProjectOrEnv,
			paths:  atProjectOrEnv("/secrets"),
			query:  &secretdto.ListSecretReq{},
			omit:   []string{argKind},
			params: settingListParams(nil),
			answer: func() any { return &secretdto.ListSecretResp{} },
		},
		{
			name: "list_registry_auths", title: "List registry credentials",
			description: "GET /projects/{project}[/{env}]/registry-auth. The credentials apps pull private " +
				"images with, each with its registry and user, its password masked: the id to give an " +
				"app's deployment settings as imageSource.registryAuth. " + descProjectOrEnv,
			paths:  atProjectOrEnv("/registry-auth"),
			query:  &registryauthdto.ListRegistryAuthReq{},
			omit:   []string{argKind},
			params: settingListParams(nil),
			answer: func() any { return &registryauthdto.ListRegistryAuthResp{} },
		},
		{
			name: "list_git_credentials", title: "List Git credentials",
			description: "GET /projects/{project}[/{env}]/git-credentials. The credentials apps clone " +
				"private repositories with - tokens, SSH keys, GitHub apps - by name: the id to give an " +
				"app's deployment settings as repoSource.credentials. " + descProjectOrEnv,
			paths:  atProjectOrEnv("/git-credentials"),
			query:  &gitcredentialdto.ListGitCredentialReq{},
			omit:   []string{argKind},
			params: settingListParams(nil),
			answer: func() any { return &gitcredentialdto.ListGitCredentialResp{} },
		},
		{
			name: "list_backup_repos", title: "List backup repositories",
			description: "GET /projects/{project}[/{env}]/backup-repos. Where app backups are kept: each " +
				"repository's engine, the storage or volume behind it, and how long backups are kept. " +
				"list_backup_snapshots lists what is in them. " + descProjectOrEnv,
			paths:  atProjectOrEnv("/backup-repos"),
			query:  &backuprepodto.ListBackupRepoReq{},
			omit:   []string{argKind},
			params: settingListParams(nil),
			answer: func() any { return &backuprepodto.ListBackupRepoResp{} },
		},
		{
			name: "list_backup_snapshots", title: "List backups",
			description: "GET /projects/{project}[/{env}]/backup-snapshots. The backups taken, newest " +
				"first: each one's repository, app, time, size and tags. Restoring one is the dashboard's " +
				"to do. " + descProjectOrEnv,
			paths: atProjectOrEnv("/backup-snapshots"),
			query: &backupsnapshotdto.ListBackupSnapshotReq{},
			params: pagingParams(map[string]string{
				"repo":        "only the backups in these repositories, by id, from list_backup_repos",
				"app":         "only the backups of these apps, by id, from list_apps",
				"tag":         "only the backups with all these tags, each key:value",
				paramFromDate: "only the backups taken on or after this day, YYYY-MM-DD",
				paramToDate:   "only the backups taken on or before this day, YYYY-MM-DD",
				paramSearch:   "a backup's short id, or text in its description",
			}),
			answer: func() any { return &backupsnapshotdto.ListBackupSnapshotResp{} },
		},
		{
			name: "list_audit_logs", title: "List audit logs",
			description: "GET /projects/{project}[/{env}]/audit-logs. Who did what, newest first: each " +
				"action's type, the user or API key that did it, what it changed, and whether it was allowed. " +
				"Where something worked yesterday and not today, this says what changed in between. Given an " +
				"env, its actions and its apps'; given only the project, the project's, its envs' and apps'.",
			paths: atProjectOrEnv("/audit-logs"),
			query: &auditlogdto.ListAuditLogReq{},
			omit:  []string{"projectId", "projectEnvId", "source", "section", "actorId", "resourceId"},
			params: pagingParams(map[string]string{
				"scopeOnly": "true for the project's or env's own actions only, not its apps'",
				"appId":     "only the actions on this app, by id, from list_apps",
				"type": "only actions of these types, such as app-update, app-create, setting-update, " +
					"project-update",
				"result":      "only actions with this result: allowed or denied",
				paramFromDate: "only actions on or after this day, YYYY-MM-DD",
				paramToDate:   "only actions on or before this day, YYYY-MM-DD",
				paramSearch:   "text in the action",
			}),
			answer: func() any { return &auditlogdto.ListAuditLogResp{} },
		},
		{
			name: "list_networks", title: "List networks",
			description: "GET /projects/{project}[/{env}]/cluster-networks. The networks apps may join " +
				"besides their env's own, each with its driver and scope. " + descProjectOrEnv,
			paths:  atProjectOrEnv("/cluster-networks"),
			query:  &networkdto.ListNetworkReq{},
			omit:   []string{argKind},
			params: settingListParams(nil),
			answer: func() any { return &networkdto.ListNetworkResp{} },
		},
		{
			name: "list_notifications", title: "List notifications",
			description: "GET /projects/{project}[/{env}]/notifications. Where deployment and failure " +
				"notices go: each notification's channels - email, Slack, Discord, Telegram, Lark - and " +
				"who gets them. " + descProjectOrEnv,
			paths:  atProjectOrEnv("/notifications"),
			query:  &notificationdto.ListNotificationReq{},
			omit:   []string{argKind},
			params: settingListParams(nil),
			answer: func() any { return &notificationdto.ListNotificationResp{} },
		},
	}
}

// ProjectEndpoints are the paths the project and env read tools use, under a
// project or an env, so that the server's tests can find each in its router.
func ProjectEndpoints() (underProjectPaths, underEnvPaths []string) {
	for _, e := range projectEndpoints() {
		underProjectPaths = append(underProjectPaths, e.paths[underProject])
		underEnvPaths = append(underEnvPaths, e.paths[underEnv])
	}
	return underProjectPaths, underEnvPaths
}
