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

// The managed template of engine mariadb, as the store lists it.
const MARIADB_SUMMARY = {
    ...SUMMARY,
    name: "mariadb-managed",
    title: "MariaDB (managed)",
    tagline: "MariaDB that HivePaaS archives continuously and restores to any moment",
    versions: [{ name: "11.8", release: "11.8.9", default: true, deprecated: false, variants: [] }],
    managed: "mariadb",
};

const MARIADB_DETAIL = { ...DETAIL, ...MARIADB_SUMMARY, license: "GPL-2.0" };

// answerTemplates answers the store's list and the managed template, and the
// lists its create form chooses an archive from: one S3 repository.
async function answerTemplates(page: Page, projectId: string, summary = SUMMARY, detail = DETAIL): Promise<void> {
    await page.route(
        url => url.pathname === "/api/app-templates",
        route => route.fulfill({ json: { meta: { page: { offset: 0, limit: 50, total: 1 } }, data: [summary] } }),
    );
    await page.route(
        url => url.pathname === `/api/app-templates/${summary.name}`,
        route => route.fulfill({ json: { meta: {}, data: detail } }),
    );
    await page.route(
        url => url.pathname === `/api/app-templates/${summary.name}/image-tags`,
        route =>
            route.fulfill({
                json: {
                    meta: {},
                    data: {
                        repository: summary.managed,
                        currentTag: summary.versions[0]?.release ?? "",
                        truncated: false,
                        tags: [],
                    },
                },
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
        capabilities: { replicas: true, verifierApp: true },
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
        target: { kind: "latest", time: null, point: "", position: "", backup: "" },
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
            pendingBytes: 48 * 1024 * 1024,
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
                position: "0/2A000028",
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
                position: "0/2B000028",
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
                    {
                        ...settingRow("e2etmpfsvol01", "e2e-in-memory", "cluster-volume"),
                        driver: "local",
                        nodeId: "e2enode01",
                        tmpfsOptions: { size: "64m" },
                    },
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

// A MariaDB template is a managed template like Postgres's: the store marks
// it, and its create form speaks of the engine's binary logs and dump tool.
test("a MariaDB managed template is created with its engine's words", async ({ page, api }) => {
    const project = await systemProject(api);
    await answerTemplates(page, project.id, MARIADB_SUMMARY, MARIADB_DETAIL);
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
                        app: { id: "e2emanagedapp002" },
                        deployment: { id: "e2edeployment003" },
                        managed: { placement: "off-site" },
                    },
                },
            });
        },
    );

    await page.goto(`/projects/${project.id}/app-templates/`);
    const card = page.getByRole("group", { name: MARIADB_SUMMARY.title });
    const badge = card.getByText("Managed", { exact: true });
    await expect(badge).toBeVisible();
    await badge.hover();
    await expect(page.getByRole("tooltip")).toContainText("HivePaaS looks after this MariaDB");
    await card.getByRole("button", { name: "Deploy" }).click();

    const dialog = page.getByRole("dialog");
    await expect(dialog.getByText("Archive and backups")).toBeVisible();
    await expect(dialog.getByText("a quiet database's binary logs are archived at least this often")).toBeVisible();
    await expect(dialog.getByText("A mariadb-dump into a backup repository")).toBeVisible();
    await expect(dialog.getByPlaceholder("SELECT COUNT(*) FROM orders")).toBeVisible();
    await dialog.getByPlaceholder("e.g. my-app").fill("e2e-mariadb");
    await dialog.getByRole("combobox", { name: "Archive to" }).click();
    await page.getByRole("option", { name: `${REPO.name} (S3)` }).click();
    await dialog.getByRole("button", { name: "Deploy Template", exact: true }).click();

    await expect.poll(() => created, { message: "the create request was sent" }).toBeDefined();
    expect(created).toMatchObject({
        name: "e2e-mariadb",
        template: MARIADB_SUMMARY.name,
        managed: { storage: { backupRepo: { id: REPO.id } }, dump: { repo: { id: REPO.id } } },
    });
});

// A MariaDB's Database tab speaks of its binary logs and its dump tool, and
// offers no replicas when its engine has none.
test("a MariaDB's Database tab uses its engine's words, and has no replicas", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    await answerManaged(page, app, {
        engine: "mariadb",
        capabilities: { replicas: false, verifierApp: true },
        locks: {
            image: "mariadb:11.8.9-noble",
            imageLine: "11.8",
            command: "mariadbd --datadir=/var/lib/mysql/data --log-bin=mariadb-bin",
            mounts: ["/var/lib/mysql"],
            envKeys: ["HIVEPAAS_ARCHIVE_TIMEOUT"],
            settings: [],
        },
    });

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    await expect(page.getByText(/^MariaDB, archived continuously/)).toBeVisible();
    const archiving = page.getByRole("region", { name: "Archiving" });
    await expect(archiving.getByText("Binary logs on the node")).toBeVisible();
    await expect(page.getByRole("region", { name: "Replicas" })).toHaveCount(0);
    await expect(page.getByText("Base backups taken on")).toHaveCount(0);
    await expect(page.getByText(/A mariadb-dump into a backup repository/)).toBeVisible();
    await expect(page.getByText(/PostgreSQL|pg_dump|\bWAL\b/)).toHaveCount(0);
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
    // A volume on every node may hold the new app's data, as the archive is off-site; one in memory never.
    await expect(page.getByRole("option", { name: /^e2e-unpinned/ })).toBeEnabled();
    await expect(page.getByRole("option", { name: /^e2e-unpinned/ })).toContainText("every node");
    await expect(page.getByRole("option", { name: /^e2e-in-memory/ })).toBeDisabled();
    await expect(page.getByRole("option", { name: /^e2e-in-memory/ })).toContainText("kept in memory");
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

// A MariaDB cluster's replicas name the host its clients reach the writer by:
// the cluster's alias, not a multi-host URL; and its replica is promoted.
test("a MariaDB's replicas name its writer's host, and one is promoted", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const { replicas } = await answerManaged(
        page,
        app,
        { engine: "mariadb", capabilities: { replicas: true, verifierApp: true } },
        { members: [REPLICA_STATUS] },
    );

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/database/`);
    const section = page.getByRole("region", { name: "Replicas" });
    await expect(section).toContainText(".HIVEPAAS_DATABASE_HOST_RW}");
    await expect(section).toContainText(".HIVEPAAS_DATABASE_HOSTS_RO}");
    await expect(section).not.toContainText("HIVEPAAS_DATABASE_URL_RW");
    await section.getByRole("button", { name: "Promote e2e-db-r1" }).click();
    await expect(page.getByText("Apps reach it through HIVEPAAS_DATABASE_HOST_RW")).toBeVisible();
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

// A database a replica was promoted over holds its cluster: it is deleted only
// with all of it, which its Danger zone asks to be confirmed, and says what goes.
test("a fenced managed database is deleted only with its whole cluster", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    await answerManaged(page, app, {
        writer: { id: "e2ereplicaapp01" },
        fencedAt: at(-HOUR),
        members: [{ ...REPLICA_STATUS, role: "primary", volume: { id: VOLUME.id }, at: at(-DAY) }],
    });
    const deletes: Request[] = [];
    await page.route(
        url => url.pathname === `/api/projects/${app.projectId}/${app.env}/apps/${app.id}`,
        async route => {
            if (route.request().method() !== "DELETE") return route.fallback();
            deletes.push(route.request());
            return route.fulfill({ json: { meta: {}, data: {} } });
        },
    );

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/danger-zone/`);
    await page.getByRole("button", { name: "Delete App" }).click();
    const dialog = page.getByRole("dialog", { name: "Delete app" });
    await expect(dialog.getByText("Delete the whole cluster")).toBeVisible();
    await expect(dialog).toContainText("e2e-db-r1 (the primary)");
    await dialog.getByRole("textbox").fill(app.name);
    await dialog.getByRole("button", { name: "Delete this App" }).click();
    await expect(dialog.getByText("it is deleted only with all of it")).toBeVisible();
    expect(deletes, "nothing is deleted unconfirmed").toHaveLength(0);

    await dialog.getByRole("checkbox", { name: /Delete the whole cluster/ }).click();
    await dialog.getByRole("button", { name: "Delete this App" }).click();
    await expect.poll(() => deletes.length, { message: "the cluster's deletion was asked for" }).toBe(1);
    const asked = new URL(deletes[0]?.url() ?? "");
    expect(asked.searchParams.get("deleteCluster")).toBe("true");
    expect(asked.searchParams.get("removeStorage")).toBe("false");
});

const WALG_SHA = "965a7147852435bc4b54f272799b63df826b54ef65cf7c4fc688c82b96c0c4eb";

// counts is a severity summary with only HIGH findings, all fixable.
function counts(high: number) {
    const none = { total: 0, fixable: 0 };
    return { critical: none, high: { total: high, fixable: high }, medium: none, low: none, unknown: none, ignored: 0 };
}

// A managed database's Security tab lists the WAL-G binary its tasks bind beside
// its image: a managed artifact, read from the artifacts cache, its modules'
// findings among the image's.
test("a managed database's artifact is shown on its Security tab", async ({ page, api }) => {
    const project = await systemProject(api);
    const app = await systemApp(api, project, "traefik");
    const scanned = { status: "done", scannedAt: at(-HOUR), updatedAt: at(-HOUR), trivyVersion: "0.75.0" };
    await page.route(
        url => url.pathname === `/api/projects/${project.id}/${app.env}/apps/${app.id}/image-scan`,
        route =>
            route.fulfill({
                json: {
                    meta: {},
                    data: {
                        enabled: true,
                        scannedAt: at(-HOUR),
                        summary: counts(1),
                        images: [
                            {
                                ...scanned,
                                imageId: `sha256:${"1".repeat(64)}`,
                                imageRef: "postgres:18.6-trixie",
                                platform: "linux/amd64",
                                osFamily: "debian",
                                osName: "13.1",
                                osEosl: false,
                                source: "node",
                                summary: counts(0),
                            },
                            {
                                ...scanned,
                                imageId: `sha256:${WALG_SHA}`,
                                imageRef: "wal-g v3.0.8",
                                platform: "linux/amd64",
                                osEosl: false,
                                source: "artifact",
                                summary: counts(1),
                            },
                        ],
                        findings: [
                            {
                                imageId: `sha256:${WALG_SHA}`,
                                vulnId: "CVE-2026-78669",
                                pkgName: "golang.org/x/net",
                                pkgType: "gobinary",
                                installedVersion: "v0.58.0",
                                fixedVersion: "0.60.0",
                                severity: "HIGH",
                            },
                        ],
                        ignored: [],
                    },
                },
            }),
    );

    await page.goto(`/projects/${project.id}/${app.env}/apps/${app.id}/security/`);
    const images = page.getByTestId("scanned-image");
    await expect(images).toHaveCount(2);
    await expect(images.nth(0)).not.toContainText("Managed artifact");
    await expect(images.nth(1)).toContainText("wal-g v3.0.8");
    await expect(images.nth(1)).toContainText("Managed artifact");
    await expect(images.nth(1)).toContainText("read from the node's artifacts cache");
    await expect(page.getByText("golang.org/x/net")).toBeVisible();
});
