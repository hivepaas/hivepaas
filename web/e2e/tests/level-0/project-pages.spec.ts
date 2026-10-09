import type { Page } from "@playwright/test";

import { systemApp, systemProject } from "../../support/api";
import { expect, test } from "../../support/fixtures";

// Every page of an app, and of a project, that takes no id beyond theirs - the
// lists, the settings, the forms to create one. They are opened on the
// HivePaaS project and its Traefik, which every installation has, and nothing
// is written: a form opened is not saved.
const APP_PAGES = [
    "general",
    "deployments",
    "instances",
    "logs",
    "metrics",
    "tasks",
    "terminal",
    "app-clone",
    "availability-and-scaling",
    "backup-snapshots",
    "config-files",
    "config-files/create",
    "container-settings",
    "danger-zone",
    "data-files",
    "data-files/create",
    "deployment-settings",
    "docker-api",
    "env-variables",
    "feature-settings",
    "kind-settings",
    "networks",
    "periodic-jobs",
    "periodic-jobs/create",
    "persistent-storage",
    "persistent-storage/create",
    "resources",
    "routing-settings",
    "sched-jobs",
    "sched-jobs/create",
    "sched-jobs/create-data-backup",
    "sched-jobs/create-sequence",
    "secrets",
    "secrets/create",
    "setting-mounts",
    "setting-mounts/create",
];

const PROJECT_PAGES = [
    "apps",
    "apps/from-compose",
    "app-templates",
    "cluster-resources/networks",
    "cluster-resources/networks/create",
    "cluster-resources/volumes",
    "cluster-resources/volumes/create",
    "integrations/access-tokens",
    "integrations/acme-dns-providers",
    "integrations/backup-repos",
    "integrations/backup-snapshots",
    "integrations/basic-auth",
    "integrations/cloud-storages",
    "integrations/command-pipes",
    "integrations/command-pipes/create",
    "integrations/command-templates",
    "integrations/command-templates/create",
    "integrations/config-files",
    "integrations/email-accounts",
    "integrations/env-variables",
    "integrations/github-apps",
    "integrations/im-platforms",
    "integrations/key-auth",
    "integrations/notification-targets",
    "integrations/registry-auth",
    "integrations/sched-jobs",
    "integrations/secrets",
    "integrations/ssh-keys",
    "integrations/ssl-certificates",
    "integrations/ssl-providers",
    "integrations/webhooks",
    "operations/audit-logs",
    "operations/export",
    "operations/tasks",
    "settings/build-settings",
    "settings/danger-zone",
    "settings/domain-settings",
    "settings/general",
];

// opensCleanly opens the page and checks nothing went wrong on the way: no
// error toast, no error in the console, no API call refused or failed, and the
// page stays at its address.
async function opensCleanly(page: Page, path: string, shows: string): Promise<void> {
    const problems: string[] = [];
    page.on("console", msg => {
        if (msg.type() === "error") problems.push(`console: ${msg.text()}`);
    });
    page.on("response", res => {
        if (res.url().includes("/api/") && res.status() >= 400) {
            problems.push(`${res.status()} ${res.request().method()} ${res.url()}`);
        }
    });

    await page.goto(path);
    await expect(page.getByText(shows, { exact: true }).first()).toBeVisible();
    await page.waitForLoadState("networkidle");
    expect(new URL(page.url()).pathname).toBe(path);
    await expect(page.locator('[data-sonner-toast][data-type="error"]')).toHaveCount(0);
    expect(problems).toEqual([]);
}

for (const tab of APP_PAGES) {
    test(`an app's ${tab} opens cleanly`, async ({ page, api }) => {
        const app = await systemApp(api, await systemProject(api), "traefik");
        await opensCleanly(page, `/projects/${app.projectId}/${app.env}/apps/${app.id}/${tab}/`, app.name);
    });
}

for (const tab of PROJECT_PAGES) {
    test(`a project's ${tab} opens cleanly`, async ({ page, api }) => {
        const project = await systemProject(api);
        await opensCleanly(page, `/projects/${project.id}/${tab}/`, "HivePaaS");
    });
}
