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
const OTHER_REPO = { id: "e2edumprepo00001", name: "e2e-dumps" };
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
        dump: { repo: { id: REPO.id }, schedule: "0 3 * * *" },
        locks: {
            image: "postgres:18.6",
            imageLine: "18",
            command: "postgres -c archive_mode=on",
            mounts: ["/var/lib/postgresql"],
            envKeys: ["PGDATA"],
            settings: [],
        },
        jobs: {
            nightly: { id: "e2ejobnightly01" },
            health: { id: "e2ejobhealth001" },
            restoreTest: null,
            dump: { id: "e2ejobdump00001" },
        },
        restore: null,
    };
}

// What switching to the database would repoint: the source it was restored
// from, and two apps that use it.
function managedSwitch() {
    return {
        source: { id: "e2esourcedb0001", name: "shop-db", key: "shop-db" },
        consumers: [
            {
                app: { id: "e2ewebapp000001", name: "shop-web", key: "shop-web" },
                envVars: ["DATABASE_URL"],
                previewClone: true,
                deployRestorePoint: false,
            },
            {
                app: { id: "e2eworkerapp001", name: "shop-worker", key: "shop-worker" },
                envVars: ["PGHOST", "PGPORT"],
                previewClone: false,
                deployRestorePoint: true,
            },
        ],
        switchedAt: null,
    };
}

// A restored database, ready.
const RESTORED = {
    restore: {
        sourceApp: { id: "e2esourcedb0001" },
        target: { kind: "latest", time: null, point: "", lsn: "", backup: "" },
        readyAt: at(-HOUR),
        switchedAt: null,
    },
};

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
        lastDump: { succeeded: true, endedAt: at(-5 * HOUR), snapshotId: "e2esnapshot0001", bytes: 3 * 2 ** 20 },
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
    settings: Record<string, unknown> = {},
    status: Record<string, unknown> = {},
): Promise<{ restores: Request[]; updates: Request[]; switches: Request[]; replicas: Request[] }> {
    const appPath = `/api/projects/${app.projectId}/${app.env}/apps/${app.id}`;
    const restores: Request[] = [];
    const updates: Request[] = [];
    const switches: Request[] = [];
    const replicas: Request[] = [];
    await page.route(
        url => url.pathname === `${appPath}/managed/promote`,
        route => {
            replicas.push(route.request());
            return route.fulfill({
                json: {
                    meta: {},
                    data: { writer: { id: "e2ereplicaapp01" }, fenced: { id: app.id }, restarted: [] },
                },
            });
        },
    );
    await page.route(
        url => url.pathname.startsWith(`${appPath}/managed/replicas`),
        route => {
            replicas.push(route.request());
            if (route.request().method() === "DELETE") return route.fulfill({ json: { meta: {} } });
            return route.fulfill({
                status: 201,
                json: { meta: {}, data: { app: { id: "e2ereplicaapp02" }, task: { id: "e2ereplicatask2" } } },
            });
        },
    );
    await page.route(
        url => url.pathname === `${appPath}/managed/switch`,
        route => {
            if (route.request().method() === "POST") {
                switches.push(route.request());
                return route.fulfill({
                    json: {
                        meta: {},
                        data: {
                            switched: [{ id: "e2eworkerapp001" }],
                            deployments: [
                                {
                                    app: { id: "e2eworkerapp001" },
                                    deployment: { id: "e2edeployment002" },
                                    task: { id: "e2edeploytask002" },
                                },
                            ],
                        },
                    },
                });
            }
            return route.fulfill({ json: { meta: {}, data: managedSwitch() } });
        },
    );
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
            if (route.request().method() === "PUT") {
                updates.push(route.request());
                return route.fulfill({ json: { meta: {} } });
            }
            if (route.request().method() !== "GET") return route.abort();
            return route.fulfill({ json: { meta: {}, data: { ...managedSettings(), ...settings } } });
        },
    );
    // What the environment can dump into.
    await page.route(
        url => url.pathname === `/api/projects/${app.projectId}/${app.env}/backup-repos`,
        route =>
            route.fulfill({
                json: paged([
                    { ...settingRow(REPO.id, REPO.name, "backup-repo"), cloudStorage: { id: "cs1" } },
                    { ...settingRow(OTHER_REPO.id, OTHER_REPO.name, "backup-repo"), volume: { id: VOLUME.id } },
                ]),
            }),
    );
    await page.route(
        url => url.pathname === `${appPath}/managed/status`,
        route => route.fulfill({ json: { meta: {}, data: { ...managedStatus(), ...status } } }),
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
    return { restores, updates, switches, replicas };
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
        // Dumped into the archive's repository, unless asked otherwise.
        managed: { storage: { backupRepo: { id: REPO.id } }, dump: { repo: { id: REPO.id } } },
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

    const lastDump = page.getByRole("region", { name: "Last dump" });
    await expect(lastDump.getByText("Succeeded")).toBeVisible();
    await expect(lastDump.getByText("3MB")).toBeVisible();
    await expect(lastDump.getByRole("link", { name: "Backup snapshots" })).toHaveAttribute(
        "href",
        new RegExp(`/apps/${app.id}/backup-snapshots/?$`),
    );

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

// The dump moves, runs at another time, or stops, from the Database tab's
// settings; a schedule the base backup's is noted.
test("a managed database's dump is changed from its settings", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const { updates } = await answerManaged(page, app);

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const settings = page.getByRole("region", { name: "Settings" });
    await expect(settings.getByRole("combobox", { name: "Dump into" })).toHaveText(REPO.name);
    const schedule = settings.getByRole("textbox", { name: "Dump schedule" });
    await expect(schedule).toHaveValue("0 3 * * *");

    await schedule.fill("0 2 * * *");
    await expect(settings.getByText("The same as the base backup's")).toBeVisible();
    await schedule.fill("0 */6 * * *");
    await expect(settings.getByText("The same as the base backup's")).toBeHidden();
    await settings.getByRole("combobox", { name: "Dump into" }).click();
    await page.getByRole("option", { name: OTHER_REPO.name }).click();
    await settings.getByRole("button", { name: "Save" }).click();

    await expect.poll(() => updates.length, { message: "the settings were saved" }).toBe(1);
    expect(updates[0]?.postDataJSON()).toEqual({
        updateVer: 3,
        dumpRepo: { id: OTHER_REPO.id },
        dumpSchedule: "0 */6 * * *",
    });
});

test("a managed database's dumps are turned off from its settings", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const { updates } = await answerManaged(page, app);

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const settings = page.getByRole("region", { name: "Settings" });
    await settings.getByRole("combobox", { name: "Dump into" }).click();
    await page.getByRole("option", { name: "No dumps" }).click();
    await settings.getByRole("button", { name: "Save" }).click();

    await expect.poll(() => updates.length, { message: "the settings were saved" }).toBe(1);
    expect(updates[0]?.postDataJSON()).toEqual({ updateVer: 3, dumpRepo: { id: "" } });
});

// A restored database lists the apps that use its source; the ones kept checked
// are switched, redeployed when asked, once the switch is confirmed.
test("a restored database switches its source's consumers to it", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const { switches } = await answerManaged(page, app, RESTORED);

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const section = page.getByRole("region", { name: "Switch consumers" });
    await expect(section.getByText("variable DATABASE_URL; its previews clone it")).toBeVisible();
    await expect(
        section.getByText("variables PGHOST, PGPORT; its deployments mark a restore point on it"),
    ).toBeVisible();

    await section.getByRole("checkbox", { name: "shop-web" }).click();
    await section.getByRole("checkbox", { name: "Redeploy them" }).click();
    await section.getByRole("button", { name: "Switch" }).click();
    await expect(page.getByText("shop-db keeps running, its jobs paused.")).toBeVisible();
    await page.getByRole("button", { name: "Switch", exact: true }).last().click();

    await expect.poll(() => switches.length, { message: "the switch was asked for" }).toBe(1);
    expect(switches[0]?.postDataJSON()).toEqual({ consumers: [{ id: "e2eworkerapp001" }], redeploy: true });
});

// A database whose consumers were switched says so, and that its jobs are paused.
test("a switched database says its jobs are paused", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    await answerManaged(page, app, { switchedTo: { app: { id: "e2erestoredapp01" }, at: at(-HOUR) } });

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const banner = page.getByRole("status").filter({ hasText: "Its consumers were switched to" });
    await expect(banner).toContainText("are paused");
    await expect(banner.getByRole("link", { name: "the database restored from it" })).toHaveAttribute(
        "href",
        /\/apps\/e2erestoredapp01\/database\/?$/,
    );
    await expect(page.getByRole("region", { name: "Switch consumers" })).toHaveCount(0);
});

// A database's replicas say how far behind they are; one is added on a pinned
// volume, and one removed once confirmed.
test("a managed database's replicas are listed, added and removed", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const { replicas } = await answerManaged(
        page,
        app,
        {},
        {
            members: [
                {
                    app: { id: "e2ereplicaapp01" },
                    key: "e2e-db-r1",
                    role: "replica",
                    node: "e2enode02",
                    state: "streaming",
                    lagBytes: 8192,
                    lagSeconds: 0.4,
                    checkedAt: at(-60_000),
                },
            ],
        },
    );

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const section = page.getByRole("region", { name: "Replicas" });
    const row = section.getByRole("row", { name: "e2e-db-r1" });
    await expect(row).toContainText("streaming");
    await expect(row).toContainText("On node e2enode02");
    await expect(row).toContainText("8KB behind, 0.4s");

    await section.getByRole("combobox", { name: "Replica volume" }).click();
    await page.getByRole("option", { name: VOLUME.name }).click();
    await section.getByRole("button", { name: "Add replica" }).click();
    await expect.poll(() => replicas.length, { message: "the replica was asked for" }).toBe(1);
    expect(replicas[0]?.method()).toBe("POST");
    expect(replicas[0]?.postDataJSON()).toEqual({ dataVolume: { id: VOLUME.id } });

    await row.getByRole("button", { name: "Remove e2e-db-r1" }).click();
    await page.getByRole("button", { name: "Remove", exact: true }).last().click();
    await expect.poll(() => replicas.length, { message: "the removal was asked for" }).toBe(2);
    expect(replicas[1]?.method()).toBe("DELETE");
    expect(new URL(replicas[1]?.url() ?? "").pathname).toMatch(/\/managed\/replicas\/e2ereplicaapp01$/);
});

// A replica's tab names its primary, and offers nothing that is the primary's.
test("a replica's Database tab names its primary", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    await answerManaged(page, app, { memberOf: { app: { id: "e2eprimaryapp01" }, role: "replica" } });

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const banner = page.getByRole("status").filter({ hasText: "A read replica of" });
    await expect(banner.getByRole("link", { name: "its primary" })).toHaveAttribute(
        "href",
        /\/apps\/e2eprimaryapp01\/database\/?$/,
    );
    await expect(page.getByRole("button", { name: "Restore" })).toHaveCount(0);
    await expect(page.getByRole("region", { name: "Settings" })).toHaveCount(0);
});

// A database with a replica can have its base backups taken there; the card says
// where they run, and taking them home again is a save away.
test("a managed database's base backups are taken on its replica", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const member = {
        app: { id: "e2ereplicaapp01" },
        key: "e2e-db-r1",
        role: "replica",
        volume: { id: VOLUME.id },
        node: "e2enode02",
        at: at(-DAY),
    };
    const { updates } = await answerManaged(page, app, { members: [member], backupFrom: { id: "e2ereplicaapp01" } });

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    await expect(page.getByRole("region", { name: "Last base backup" })).toContainText("Runs one2e-db-r1");
    const settings = page.getByRole("region", { name: "Settings" });
    const host = settings.getByRole("combobox", { name: "Base backups taken on" });
    await expect(host).toHaveText("Replica e2e-db-r1");

    await host.click();
    await page.getByRole("option", { name: "This database" }).click();
    await settings.getByRole("button", { name: "Save" }).click();
    await expect.poll(() => updates.length, { message: "the settings were saved" }).toBe(1);
    expect(updates[0]?.postDataJSON()).toEqual({ updateVer: 3, backupFrom: { id: "" } });
});

const REPLICA_STATUS = {
    app: { id: "e2ereplicaapp01" },
    key: "e2e-db-r1",
    role: "replica",
    node: "e2enode02",
    state: "streaming",
    lagBytes: 0,
    lagSeconds: null,
    checkedAt: at(-60_000),
};

// A replica is promoted once confirmed, the confirmation saying what stops; the
// section says where apps reach the primary.
test("a managed database's replica is promoted", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const { replicas } = await answerManaged(page, app, {}, { members: [REPLICA_STATUS] });

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const section = page.getByRole("region", { name: "Replicas" });
    await expect(section).toContainText(".HIVEPAAS_DATABASE_URL_RW}");
    await section.getByRole("button", { name: "Promote e2e-db-r1" }).click();
    await expect(page.getByText("is stopped, and kept from starting its server")).toBeVisible();
    await page.getByRole("button", { name: "Promote", exact: true }).last().click();

    await expect.poll(() => replicas.length, { message: "the promotion was asked for" }).toBe(1);
    expect(new URL(replicas[0]?.url() ?? "").pathname).toMatch(/\/managed\/promote$/);
    expect(replicas[0]?.postDataJSON()).toEqual({ member: { id: "e2ereplicaapp01" } });
});

// A database a replica was promoted over says it is fenced and who writes; the
// primary member is neither promoted nor removed.
test("a fenced managed database says which member is the primary", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const primary = { ...REPLICA_STATUS, role: "primary", state: "primary" };
    await answerManaged(
        page,
        app,
        {
            writer: { id: "e2ereplicaapp01" },
            fencedAt: at(-HOUR),
            members: [{ ...REPLICA_STATUS, role: "primary", volume: { id: VOLUME.id }, at: at(-DAY) }],
        },
        { members: [primary] },
    );

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    await expect(page.getByRole("status").filter({ hasText: "Fenced since" })).toContainText(
        "e2e-db-r1 is the cluster's primary",
    );
    const row = page.getByRole("region", { name: "Replicas" }).getByRole("row", { name: "e2e-db-r1" });
    await expect(row).toContainText("primary");
    await expect(row.getByRole("button", { name: "Promote e2e-db-r1" })).toHaveCount(0);
    await expect(row.getByRole("button", { name: "Remove e2e-db-r1" })).toBeDisabled();
});
