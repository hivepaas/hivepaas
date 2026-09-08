import { ROUTE } from "@application/shared/constants";

/**
 * One object that still references a setting, as the API reports it.
 *
 * Everything past type and id is best effort - the link table stores ids, and
 * what it points at may have been deleted since.
 */
export type SettingUsage = {
    type: string;
    id: string;
    name?: string;
    settingType?: string;
    scope?: string;
    projectId?: string;
    projectEnvKey?: string;
    appId?: string;
    appName?: string;
    userId?: string;
};

/**
 * Where a referencing object can be edited.
 *
 * The dashboard's routes are regular enough that this is mostly one table: the
 * same slug appears under /integrations for a global setting and under
 * /projects/:id/integrations for a project one, so scope picks the prefix and the
 * setting type picks the page.
 *
 * A type this does not know returns null, and the caller shows the row without a
 * link. That is deliberate: forty-odd setting types exist and more will be added,
 * and a row saying "some setting named X uses this" is far more useful than
 * either hiding it or guessing at a URL that 404s.
 */
// Route builders rather than route keys.
//
// Keying into ROUTE by name needs the $pattern and $route siblings excluded and
// then still breaks on the one node that has no edit page, so each entry names
// its builder outright. It is more lines, and every one of them is checked where
// it is written: a setting type pointed at the wrong page fails to compile here
// rather than 404ing for somebody later.
const G = ROUTE.settings;
const P = ROUTE.projects.single.providerConfiguration;
const A = ROUTE.projects.single.apps.single.configuration;

type GlobalHref = (id: string) => string;
type ProjectHref = (projectId: string, id: string) => string;
type AppHref = (projectId: string, env: string, appId: string) => string;

/** Settings that live on an /integrations page. */
const GLOBAL_HREF: Record<string, GlobalHref> = {
    "access-token": G.accessTokens.edit.$route,
    "acme-dns-provider": G.acmeDnsProviders.edit.$route,
    "backup-repo": G.backupRepos.edit.$route,
    "basic-auth": G.basicAuth.edit.$route,
    "cloud-storage": G.cloudStorages.edit.$route,
    "email": G.emailAccounts.edit.$route,
    "github-app": G.githubApps.edit.$route,
    "im-service": G.imPlatforms.edit.$route,
    "notification": G.notificationTargets.edit.$route,
    "oauth": G.oauth.edit.$route,
    "registry-auth": G.registryAuth.edit.$route,
    "repo-webhook": G.webhooks.edit.$route,
    "ssh-key": G.sshKeys.edit.$route,
    "ssl-cert": G.sslCertificates.edit.$route,
    "ssl-provider": G.sslProviders.edit.$route,
};

/** The same settings under a project, plus the few that only exist there. */
const PROJECT_HREF: Record<string, ProjectHref> = {
    "access-token": P.accessTokens.edit.$route,
    "acme-dns-provider": P.acmeDnsProviders.edit.$route,
    "backup-repo": P.backupRepos.edit.$route,
    "basic-auth": P.basicAuth.edit.$route,
    "cloud-storage": P.cloudStorages.edit.$route,
    "command-pipe": P.commandPipes.edit.$route,
    "command-template": P.commandTemplates.edit.$route,
    "email": P.emailAccounts.edit.$route,
    "github-app": P.githubApps.edit.$route,
    "im-service": P.imPlatforms.edit.$route,
    "notification": P.notificationTargets.edit.$route,
    "registry-auth": P.registryAuth.edit.$route,
    "repo-webhook": P.webhooks.edit.$route,
    "secret": P.secrets.edit.$route,
    "ssh-key": P.sshKeys.edit.$route,
    "ssl-cert": P.sslCertificates.edit.$route,
    "ssl-provider": P.sslProviders.edit.$route,
};

/** App-scoped settings are tabs on the app, not pages of their own. */
const APP_HREF: Record<string, AppHref> = {
    "app": A.general.$route,
    "app-clone": A.appClone.$route,
    "app-deployment": A.deploymentSettings.$route,
    "app-features": A.featureSettings.$route,
    "app-placement": A.availabilityAndScaling.$route,
    "app-routing": A.routingSettings.$route,
    "config-file": A.configFiles.$route,
    "env-var": A.envVariables.$route,
    "periodic-job": A.periodicJobs.$route,
    "sched-job": A.scheduledJobs.$route,
    "secret": A.secrets.$route,
};

/** Settings that live on exactly one page, whatever scope they claim. */
const SYSTEM_HREF: Record<string, string> = {
    "backup-repo-cleanup": ROUTE.systemSettings.backupRepoCleanup.configuration.$route,
    "hivepaas-service": ROUTE.systemSettings.hivepaas.general.$route,
    "image-build": ROUTE.appSettings.imageBuild.$route,
    "ssl-renewal": ROUTE.systemSettings.sslRenewal.configuration.$route,
    "system-backup": ROUTE.systemSettings.dataBackup.configuration.$route,
    "system-cleanup": ROUTE.systemSettings.dataCleanup.configuration.$route,
    "traefik-service": ROUTE.systemSettings.traefik.general.$route,
};

export function settingUsageHref(usage: SettingUsage): string | null {
    const { settingType, scope } = usage;
    if (settingType == null || settingType === "") {
        return null;
    }

    // Fixed pages first: there is only one of each, so scope adds nothing.
    const systemHref = SYSTEM_HREF[settingType];
    if (systemHref != null) {
        return systemHref;
    }

    if (scope === "app") {
        const { projectId, projectEnvKey, appId } = usage;
        // All three are needed: app routes are /projects/:id/:env/apps/:appId/...
        // and env is the project env key, not its id.
        if (projectId == null || projectEnvKey == null || appId == null) {
            return null;
        }
        return APP_HREF[settingType]?.(projectId, projectEnvKey, appId) ?? null;
    }

    if (scope === "project" || scope === "project-env") {
        const { projectId } = usage;
        if (projectId == null || projectId === "") {
            return null;
        }
        if (settingType === "domain-settings") {
            return ROUTE.projects.single.configuration.domainSettings.$route(projectId);
        }
        return PROJECT_HREF[settingType]?.(projectId, usage.id) ?? null;
    }

    // Global and hivepaas scope share the /integrations pages.
    return GLOBAL_HREF[settingType]?.(usage.id) ?? null;
}

/** What to call the object in the list, when the API could not name it. */
export function settingUsageLabel(usage: SettingUsage): string {
    if (usage.name != null && usage.name !== "") {
        return usage.appName != null && usage.appName !== "" ? `${usage.name} (${usage.appName})` : usage.name;
    }
    return usage.settingType != null && usage.settingType !== "" ? usage.settingType : usage.id;
}
