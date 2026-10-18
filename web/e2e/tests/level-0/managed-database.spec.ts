import type { Page, Request } from "@playwright/test";

import { systemApp, systemProject } from "../../support/api";
import { expect, test } from "../../support/fixtures";

// A managed database as the dashboard shows and steers it: created from a
// managed template with the archive chosen; its Database tab's status; and a
// restore into a new app. The local backend has no managed database, so the
// managed template, the database's engine and the lists its forms choose from
// are answered by the test (page.route), as license-display.spec.ts answers the
// license: the store is browsed in the HivePaaS project, and its Traefik app
// stands in for the database. Every write is answered here and checked for what
// it would have sent: nothing reaches the server.

const HOUR = 3_600_000;
const DAY = 24 * HOUR;

// at is a moment from now, as the API writes it.
function at(fromNow: number): string {
    return new Date(Date.now() + fromNow).toISOString();
}

const TEMPLATE = "postgres-managed";
const REPO = { id: "e2emanagedrepo01", name: "e2e-archive" };
const VOLUME = { id: "e2emanagedvol001", name: "e2e-restore-data" };
const POINT = "before-migration";
const DEPLOY_POINT = "deploy-shop-web-20261018093015";

// A setting's fields every list's rows carry.
function settingRow(id: string, name: string, type: string) {
    return { id, name, status: "active", type, updateVer: 1, createdAt: at(-DAY) };
}

function paged(data: unknown[]) {
    return { meta: { page: { limit: 1000, offset: 0, total: data.length } }, data };
}

const SUMMARY = {
    name: TEMPLATE,
    title: "PostgreSQL (managed)",
    tagline: "PostgreSQL, archived continuously and restorable to any moment",
    categories: ["databases"],
    tags: ["sql"],
    aliases: [],
    iconUrl: "",
    variants: [],
    versions: [{ name: "18", release: "18.6", default: true, deprecated: false, variants: [] }],
    compatible: true,
    managed: "postgres",
    requiresFeatures: ["managed-databases"],
    missingFeatures: [],
};

const DETAIL = {
    ...SUMMARY,
    source: "e2e",
    revision: "e2e",
    description: "",
    license: "PostgreSQL",
    managedDefaults: {
        archiveTimeout: "60s",
        baseBackupSchedule: "0 2 * * *",
        retainFull: 7,
        restoreTestSchedule: "0 4 * * 0",
    },
    parameters: [],
};

// answerTemplates answers the store's list and the managed template, and the
// lists its create form chooses an archive from: one S3 repository.
async function answerTemplates(page: Page, projectId: string): Promise<void> {
    await page.route(
        url => url.pathname === "/api/app-templates",
        route => route.fulfill({ json: { meta: { page: { offset: 0, limit: 50, total: 1 } }, data: [SUMMARY] } }),
    );
    await page.route(
        url => url.pathname === `/api/app-templates/${TEMPLATE}`,
        route => route.fulfill({ json: { meta: {}, data: DETAIL } }),
    );
    await page.route(
        url => url.pathname === `/api/app-templates/${TEMPLATE}/image-tags`,
        route =>
            route.fulfill({
                json: { meta: {}, data: { repository: "postgres", currentTag: "18.6", truncated: false, tags: [] } },
            }),
    );
    // The environment's list: what the environment chosen can archive to.
    await page.route(
        url => url.pathname.startsWith(`/api/projects/${projectId}/`) && url.pathname.endsWith("/backup-repos"),
        route =>
            route.fulfill({
                json: paged([{ ...settingRow(REPO.id, REPO.name, "backup-repo"), cloudStorage: { id: "cs1" } }]),
            }),
    );
}

// The database's engine as its endpoints answer: archiving, two base backups,
// a restore point, and the settings it was made with.
function managedSettings() {
    return {
        updateVer: 3,
        engine: "postgres",
        engineApi: 1,
        storage: {
            backupRepo: { id: REPO.id },
            cloudStorage: null,
            keyAuth: null,
            prefix: "walg/e2e-db",
            placement: "off-site",
        },
        archiveTimeout: "60s",
        retainFull: 7,
        baseBackupSchedule: "0 2 * * *",
        restoreTestSchedule: "0 4 * * 0",
        restoreTest: { mode: "container", verifierApp: null, nodeId: "", keyAuth: null, checkQuery: "" },
        locks: {
            image: "postgres:18.6",
            imageLine: "18",
            command: "postgres -c archive_mode=on",
            mounts: ["/var/lib/postgresql"],
            envKeys: ["PGDATA"],
            settings: [],
        },
        jobs: { nightly: { id: "e2ejobnightly01" }, health: { id: "e2ejobhealth001" }, restoreTest: null },
        restore: null,
    };
}

function managedStatus() {
    return {
        engine: "postgres",
        engineApi: 1,
        archiving: {
            state: "ok",
            checkedAt: at(-60_000),
            lastArchivedAt: at(-90_000),
            failedCount: 0,
            pgWalBytes: 48 * 1024 * 1024,
            stale: false,
        },
        restoreWindow: { from: at(-2 * DAY), to: at(-90_000) },
        lastBackup: { succeeded: true, endedAt: at(-6 * HOUR), name: "base_000000010000000000000042", bytes: 2 ** 30 },
        backups: [
            { name: "base_000000010000000000000021", at: at(-30 * HOUR) },
            { name: "base_000000010000000000000042", at: at(-6 * HOUR) },
        ],
        lastRestoreTest: { passed: true, endedAt: at(-3 * DAY), durationSeconds: 125, restoredTo: at(-3 * DAY) },
        placement: "off-site",
        suggestions: [],
        restorePoints: [
            {
                name: POINT,
                lsn: "0/2A000028",
                at: at(-5 * HOUR),
                by: null,
                note: "Before the orders migration",
                keepUntil: null,
                kept: false,
                restorableUntil: at(5 * DAY),
                origin: "person",
            },
            {
                name: DEPLOY_POINT,
                lsn: "0/2B000028",
                at: at(-4 * HOUR),
                by: null,
                note: "Before deployment e2edeploy00001 of shop web",
                keepUntil: null,
                kept: false,
                restorableUntil: at(5 * DAY),
                origin: "deploy",
            },
        ],
        heldBack: null,
    };
}

// answerManaged makes the app a managed database: its kind says database, and
// its engine's endpoints answer. A restore is answered with a new app and its
// task, and handed to the test.
async function answerManaged(
    page: Page,
    app: { projectId: string; env: string; id: string },
): Promise<{ restores: Request[] }> {
    const appPath = `/api/projects/${app.projectId}/${app.env}/apps/${app.id}`;
    const restores: Request[] = [];
    await page.route(
        url => url.pathname === appPath,
        async route => {
            if (route.request().method() !== "GET") return route.abort();
            const response = await route.fetch();
            const body = (await response.json()) as { data: Record<string, unknown> };
            return route.fulfill({ response, json: { ...body, data: { ...body.data, category: "database" } } });
        },
    );
    await page.route(
        url => url.pathname === `${appPath}/managed`,
        route => {
            if (route.request().method() !== "GET") return route.abort();
            return route.fulfill({ json: { meta: {}, data: managedSettings() } });
        },
    );
    await page.route(
        url => url.pathname === `${appPath}/managed/status`,
        route => route.fulfill({ json: { meta: {}, data: managedStatus() } }),
    );
    await page.route(
        url => url.pathname === `${appPath}/managed/restore`,
        route => {
            restores.push(route.request());
            return route.fulfill({
                json: { meta: {}, data: { app: { id: "e2erestoredapp01" }, task: { id: "e2erestoretask01" } } },
            });
        },
    );
    await page.route(
        url => url.pathname === `/api/projects/${app.projectId}/${app.env}/cluster-volumes`,
        route =>
            route.fulfill({
                json: paged([
                    { ...settingRow(VOLUME.id, VOLUME.name, "cluster-volume"), driver: "local", nodeId: "e2enode01" },
                    { ...settingRow("e2eunpinnedvol01", "e2e-unpinned", "cluster-volume"), driver: "local" },
                ]),
            }),
    );
    return { restores };
}

// A managed template is deployed with the archive it is given: the store marks
// it managed, its form asks where it archives, and the request carries it.
test("a managed template is created with the archive chosen", async ({ page, api }) => {
    const project = await systemProject(api);
    await answerTemplates(page, project.id);
    let created: Record<string, unknown> | undefined;
    await page.route(
        url => url.pathname.endsWith("/apps/from-template/preflight"),
        route => route.fulfill({ json: { meta: {}, data: { storage: [], issues: [] } } }),
    );
    await page.route(
        url => url.pathname.endsWith("/apps/from-template"),
        route => {
            created = route.request().postDataJSON() as Record<string, unknown>;
            return route.fulfill({
                json: {
                    meta: {},
                    data: {
                        app: { id: "e2emanagedapp001" },
                        deployment: { id: "e2edeployment001" },
                        managed: { placement: "off-site" },
                    },
                },
            });
        },
    );

    await page.goto(`/projects/${project.id}/app-templates/`);
    const card = page.getByRole("group", { name: SUMMARY.title });
    await expect(card.getByText("Managed", { exact: true })).toBeVisible();
    await card.getByRole("button", { name: "Deploy" }).click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByText("Archive and backups")).toBeVisible();
    await dialog.getByPlaceholder("e.g. my-app").fill("e2e-db");
    await dialog.getByRole("combobox", { name: "Archive to" }).click();
    await page.getByRole("option", { name: `${REPO.name} (S3)` }).click();
    await dialog.getByRole("button", { name: "Deploy Template", exact: true }).click();

    await expect.poll(() => created, { message: "the create request was sent" }).toBeDefined();
    expect(created).toMatchObject({
        name: "e2e-db",
        template: TEMPLATE,
        managed: { storage: { backupRepo: { id: REPO.id } } },
    });
    await expect(page).toHaveURL(/\/apps\/e2emanagedapp001\/deployments\/?$/);
});

// A managed database's Database tab says how its archive is doing, how far
// back it can be restored, and its restore points.
test("a managed database's Database tab shows its engine's status", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    await answerManaged(page, app);

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/metrics/`);
    await page.getByRole("link", { name: "Database", exact: true }).click();
    await expect(page).toHaveURL(/\/database\/?$/);

    const archiving = page.getByRole("region", { name: "Archiving" });
    await expect(archiving.getByText("Archiving", { exact: true }).last()).toBeVisible();
    await expect(archiving.getByText("48MB")).toBeVisible();
    const restoreWindow = page.getByRole("region", { name: "Restore window" });
    await expect(restoreWindow.getByText("2 kept, of 7")).toBeVisible();
    await expect(page.getByRole("region", { name: "Last base backup" }).getByText("Succeeded")).toBeVisible();
    await expect(page.getByRole("region", { name: "Last restore test" }).getByText("2m 5s")).toBeVisible();
    await expect(page.getByText("Off-site", { exact: true })).toBeVisible();

    const points = page.getByRole("region", { name: "Restore points" });
    await expect(points.getByRole("row", { name: new RegExp(POINT) })).toContainText("Before the orders migration");
    await expect(points.getByRole("row", { name: new RegExp(POINT) })).toContainText("Not kept");
    // A deployment's point says so; a person's does not.
    const deployBadge = (name: string) =>
        points.getByRole("row", { name: new RegExp(name) }).getByText("Deploy", { exact: true });
    await expect(deployBadge(DEPLOY_POINT)).toBeVisible();
    await expect(deployBadge(POINT)).toHaveCount(0);
});

// A restore goes into a new app, to the target chosen, on the volume chosen,
// and the dialog then shows the task the new app recovers in.
test("a managed database is restored to a restore point into a new app", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const { restores } = await answerManaged(page, app);

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    await page.getByRole("button", { name: "Restore", exact: true }).click();

    const dialog = page.getByRole("dialog", { name: `Restore ${app.name}` });
    await dialog.getByRole("tab", { name: "A restore point" }).click();
    await dialog.getByRole("combobox", { name: "Restore point" }).click();
    await page.getByRole("option", { name: new RegExp(`^${POINT}`) }).click();
    await dialog.getByRole("combobox", { name: "Data volume" }).click();
    // A volume pinned to no node cannot hold the new app's data.
    await expect(page.getByRole("option", { name: /^e2e-unpinned/ })).toBeDisabled();
    await page.getByRole("option", { name: new RegExp(`^${VOLUME.name}`) }).click();
    await dialog.getByRole("button", { name: "Restore", exact: true }).click();

    await expect.poll(() => restores.length, { message: "the restore was asked for" }).toBe(1);
    expect(restores[0]?.postDataJSON()).toEqual({
        target: { kind: "point", point: POINT },
        dataVolume: { id: VOLUME.id },
    });
    await expect(page).toHaveURL(/\/apps\/e2erestoredapp01\/tasks\/e2erestoretask01\/?$/);
});
