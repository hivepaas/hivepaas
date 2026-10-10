import type { APIRequestContext, Page } from "@playwright/test";
import { readFile } from "node:fs/promises";

import {
    type App,
    appPath,
    createApp,
    createJob,
    createVolume,
    created,
    deleteSettingsNamed,
    deleteVolume,
    deployImage,
    runJob,
    setRuntimeEnvVars,
    settingIdNamed,
} from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs, mountVolume } from "../../support/apps";
import {
    REPO_PASSWORD,
    confirmRestore,
    repoAction,
    snapshotsIn,
    startRestore,
    storeVolume,
    volumeBackupJob,
    volumeRepo,
} from "../../support/backups";
import { type Cleanup, e2eName, expect, test } from "../../support/fixtures";

// The database a command backs up.
const POSTGRES = "postgres:18.6-alpine";

test.describe.configure({ timeout: 300_000 });

test("a data backup of an app's volume is a snapshot, restored", async ({ page, api, cleanup }) => {
    // Made before the project, so removed after it: the app mounts them.
    const data = e2eName("backed-up-data");
    const dataId = await createVolume(api, data);
    cleanup(() => deleteVolume(api, dataId));
    const store = e2eName("backups");
    const storeId = await createVolume(api, store, `/hp/${store}`);
    cleanup(() => deleteVolume(api, storeId));
    cleanup(() => deleteSettingsNamed(api, "backup-repos", store));
    const app = await appIn(api, cleanup, "backup");

    await mountVolume(page, app, data, "/data");
    const mark = e2eName("mark");
    await deployImage(api, app, BUSYBOX, `sh -c 'echo ${mark} > /data/mark; echo ready; exec sleep 3600'`);
    await deployed(api, app);
    await expectLogs(page, app, "ready");

    await page.goto("/integrations/backup-repos/create/");
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill(store);
    await page.getByRole("group", { name: "Type" }).getByRole("tab", { name: "Volume" }).click();
    await page.getByRole("group", { name: "Volume", exact: true }).getByRole("combobox").click();
    await page.getByRole("option", { name: store }).click();
    await page.getByRole("group", { name: "Password *" }).getByRole("textbox").fill("E2e-backup-Pass1");
    await page.getByRole("group", { name: "Available in Projects" }).getByRole("checkbox").check();
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: new RegExp(store) })).toBeVisible();

    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Data Backup" }).click();
    await page.getByRole("group", { name: "Name *", exact: true }).getByRole("textbox").fill("e2e-backup");
    await page.getByRole("group", { name: "Back Up" }).getByRole("tab", { name: "Volume" }).click();
    await page.getByRole("group", { name: "Volume *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: "/data" }).click();
    await page.getByRole("group", { name: "Backup Repository *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: store }).click();
    await page.getByRole("button", { name: "Save" }).click();

    const row = page.getByRole("row", { name: /e2e-backup/ });
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: "Run Now" }).click();
    await page.getByRole("dialog", { name: "Task created" }).getByRole("button", { name: "View Run" }).click();
    await expect(page.locator("main")).toContainText(/Done\s*task:/, { timeout: 120_000 });

    // From here, each container says what it found, with a name of its own -
    // they all have the app's hostname - then changes it: one started after a
    // restore finds the mark again.
    await deployImage(
        api,
        app,
        BUSYBOX,
        "sh -c 'echo \"found-$(cat /data/mark)-in-$(cat /proc/sys/kernel/random/uuid)\"; echo changed > /data/mark; exec sleep 3600'",
    );
    await deployed(api, app);
    const before = await expectLogs(page, app, new RegExp(`found-${mark}-in-[\\w-]+`));

    await page.goto(appPage(app, "backup-snapshots"));
    const snapshot = page.getByRole("row", { name: /e2e-backup/ });
    await snapshot.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("menuitem", { name: "Restore" }).click();
    await expect(page.getByRole("heading", { name: /^Restore Snapshot/ })).toBeVisible();
    // Replace: the directory as it was backed up, the app stopped meanwhile.
    await page.getByRole("group", { name: "App Name" }).getByRole("textbox").fill(app.name);
    await page.getByRole("button", { name: "Restore", exact: true }).click();
    await expect(page.locator("main")).toContainText(/Done\s*task:/, { timeout: 120_000 });

    // The app, started again in a container of its own, finds the mark the
    // backup holds, not "changed".
    const first = before.split("-in-")[1];
    await expectLogs(page, app, new RegExp(`found-${mark}-in-(?!${first})[\\w-]+`));
});

// A database dumped by its data backup job's command is a snapshot; restored,
// the dump is fed to the job's restore command, and the database has back what
// it had when dumped. The dump reaches the command through the repository's
// server: the stream a restore reads.
test("a database dumped by a command is a snapshot, restored through the restore command", async ({
    page,
    api,
    cleanup,
}) => {
    const store = e2eName("dumps");
    const storeId = await createVolume(api, store, `/hp/${store}`);
    cleanup(() => deleteVolume(api, storeId));
    cleanup(() => deleteSettingsNamed(api, "backup-repos", store));
    const repo = await api.post("settings/backup-repos", {
        data: {
            name: store,
            engine: "kopia",
            volume: { id: storeId },
            password: "E2e-backup-Pass1",
            inheritable: true,
        },
    });
    expect(repo.ok(), `creating backup repository ${store}: ${await repo.text()}`).toBe(true);
    const app = await appIn(api, cleanup, "dump");
    await setRuntimeEnvVars(api, app, [{ key: "POSTGRES_PASSWORD", value: "E2e-pg-Pass1", isLiteral: true }]);
    await deployImage(api, app, POSTGRES);
    await deployed(api, app);
    await expectLogs(page, app, "PostgreSQL init process complete");

    const mark = e2eName("note");
    const psql = (sql: string) => `psql -U postgres -v ON_ERROR_STOP=1 -tAc "${sql}"`;
    const seed = await createJob(
        api,
        app,
        "e2e-seed",
        psql(`CREATE TABLE notes(v text); INSERT INTO notes VALUES ('${mark}')`),
    );
    await expect.poll(async () => (await runJob(api, app, seed)).status, { timeout: 90_000 }).toBe("done");

    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Data Backup" }).click();
    await page.getByRole("group", { name: "Name *", exact: true }).getByRole("textbox").fill("e2e-dump");
    await page.getByRole("group", { name: "Back Up" }).getByRole("tab", { name: "Command" }).click();
    await page.getByRole("group", { name: "File Name *" }).getByRole("textbox").fill("db.sql");
    const commands = page.getByRole("group", { name: "Command *" });
    await commands.nth(0).getByRole("textbox").fill("pg_dump -U postgres --clean --if-exists postgres");
    await commands.nth(1).getByRole("textbox").fill("psql -U postgres -v ON_ERROR_STOP=1 postgres");
    await page.getByRole("group", { name: "Backup Repository *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: store }).click();
    await page.getByRole("button", { name: "Save" }).click();
    const row = page.getByRole("row", { name: /e2e-dump/ });
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: "Run Now" }).click();
    await page.getByRole("dialog", { name: "Task created" }).getByRole("button", { name: "View Run" }).click();
    await expect(page.locator("main")).toContainText(/Done\s*task:/, { timeout: 120_000 });

    const drop = await createJob(api, app, "e2e-drop", psql("DROP TABLE notes"));
    expect((await runJob(api, app, drop)).status).toBe("done");

    await page.goto(appPage(app, "backup-snapshots"));
    await page
        .getByRole("row", { name: /e2e-dump/ })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("menuitem", { name: "Restore" }).click();
    await expect(page.getByRole("heading", { name: /^Restore Snapshot/ })).toBeVisible();
    await page.getByRole("group", { name: "App Name" }).getByRole("textbox").fill(app.name);
    await page.getByRole("button", { name: "Restore", exact: true }).click();
    await expect(page.locator("main")).toContainText(/Done\s*task:/, { timeout: 120_000 });

    const read = await createJob(api, app, "e2e-read", psql("SELECT v FROM notes"));
    const result = await runJob(api, app, read);
    expect(result.status).toBe("done");
    expect(result.log).toContain(mark);
});

// dataApp is an app of a project of its own with a volume of its own mounted
// at /data, running busybox on the command; it answers the app and the
// volume's id.
async function dataApp(
    page: Page,
    api: APIRequestContext,
    cleanup: Cleanup,
    label: string,
    command: string,
): Promise<{ app: App; volumeId: string }> {
    const data = e2eName(`${label}-data`);
    const volumeId = await createVolume(api, data);
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, label);
    await mountVolume(page, app, data, "/data");
    await deployImage(api, app, BUSYBOX, `sh -c '${command}; echo ready; exec sleep 3600'`);
    await deployed(api, app);
    await expectLogs(page, app, "ready");
    return { app, volumeId };
}

// shown runs the command in the app, as a job of its own, and answers what it
// printed.
let shows = 0;
async function shown(api: APIRequestContext, app: App, command: string): Promise<string> {
    const job = await createJob(api, app, `e2e-show-${++shows}`, `sh -c '${command}'`);
    const result = await runJob(api, app, job);
    expect(result.status, result.log).toBe("done");
    return result.log;
}

// backedUp runs the backup job, and waits for it to be done.
async function backedUp(api: APIRequestContext, app: App, job: string): Promise<void> {
    const result = await runJob(api, app, job);
    expect(result.status, result.log).toBe("done");
}

// A snapshot restored into another app of the project with Overwrite adds to
// what that app has: its files of the same names are the snapshot's, the
// others stay. The app keeps running meanwhile, as one serving static files
// would.
test("a snapshot restored into another app with Overwrite keeps that app's other files", async ({
    page,
    api,
    cleanup,
}) => {
    const repo = await volumeRepo(api, cleanup, "overwrites");
    const mark = e2eName("mark");
    const { app: source, volumeId } = await dataApp(page, api, cleanup, "restore-from", `echo ${mark} > /data/mark`);
    const job = await volumeBackupJob(api, source, "e2e-backup", volumeId, repo);
    await backedUp(api, source, job);

    const target = await createApp(api, { id: source.projectId }, "copy");
    await mountVolume(page, target, e2eName("restore-from-data"), "/data");
    await deployImage(api, target, BUSYBOX, "sh -c 'echo own > /data/mark; echo kept > /data/extra; exec sleep 3600'");
    await deployed(api, target);
    await expect.poll(() => shown(api, target, "cat /data/mark"), { timeout: 60_000 }).toContain("own");

    await startRestore(page, `/projects/${source.projectId}/integrations/backup-snapshots/`, "e2e-backup");
    await page
        .getByRole("group", { name: /^Restore Into/ })
        .getByRole("combobox")
        .click();
    await page.getByRole("option").filter({ hasText: "copy" }).click();
    await expect(page.getByText(`The data of ${source.name} is loaded into copy.`)).toBeVisible();
    await page
        .getByRole("group", { name: /^Volume/ })
        .getByRole("combobox")
        .click();
    await page.getByRole("option", { name: "/data" }).click();
    await page.getByRole("tab", { name: "Overwrite" }).click();
    await page.getByLabel("Stop the app while restoring").uncheck();
    await confirmRestore(page, "copy");

    const files = await shown(api, target, "cat /data/mark /data/extra");
    expect(files).toContain(mark);
    expect(files).toContain("kept");
    expect(files).not.toContain("own");
});

// One directory of a snapshot restored with Replace is that directory as it
// was backed up, its later files moved aside with it; the snapshot's other
// directories are not touched.
test("a directory of a snapshot restored with Replace is as it was, the rest untouched", async ({
    page,
    api,
    cleanup,
}) => {
    const repo = await volumeRepo(api, cleanup, "replaces");
    const { app, volumeId } = await dataApp(
        page,
        api,
        cleanup,
        "replace",
        "mkdir -p /data/a /data/b; [ -f /data/a/x ] || { echo one > /data/a/x; echo one > /data/b/y; }",
    );
    const job = await volumeBackupJob(api, app, "e2e-backup", volumeId, repo);
    await backedUp(api, app, job);
    await shown(api, app, "echo two > /data/a/x; echo new > /data/a/new; echo two > /data/b/y");

    await startRestore(page, appPage(app, "backup-snapshots"), "e2e-backup");
    await page.getByRole("radio", { name: "Restore a", exact: true }).check();
    await confirmRestore(page, app.name);

    await expect
        .poll(() => shown(api, app, "cat /data/a/x /data/b/y; ls /data/a /data/a.before-restore-*"), {
            timeout: 60_000,
        })
        .toMatch(/one\s+two\s+\/data\/a:\s+x\s+\/data\/a\.before-restore-[\d-]+:\s+new\s+x/);
});

// A snapshot's files are read from its details, one downloaded; deleted, the
// snapshot is gone from the repository, and a sync of the repository does not
// bring it back.
test("a snapshot's file is downloaded, and the snapshot deleted stays so after a sync", async ({
    page,
    api,
    cleanup,
}) => {
    const repo = await volumeRepo(api, cleanup, "deletes");
    const mark = e2eName("note");
    const { app, volumeId } = await dataApp(
        page,
        api,
        cleanup,
        "download",
        `mkdir -p /data/notes; echo ${mark} > /data/notes/today`,
    );
    const job = await volumeBackupJob(api, app, "e2e-backup", volumeId, repo);
    await backedUp(api, app, job);

    await page.goto(appPage(app, "backup-snapshots"));
    const row = page.getByRole("row", { name: /e2e-backup/ });
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("menuitem", { name: "View details" }).click();
    const details = page.getByRole("dialog");
    await details.getByRole("button", { name: "notes/" }).click();
    const download = page.waitForEvent("download");
    await details.getByRole("button", { name: "Download notes/today" }).click();
    const file = await (await download).path();
    expect((await readFile(file, "utf8")).trim()).toBe(mark);
    await page.keyboard.press("Escape");

    const [snapshot] = await snapshotsIn(api, repo);
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("menuitem", { name: "Delete" }).click();
    const dialog = page.getByRole("dialog", { name: "Delete Backup Snapshot" });
    await dialog.getByRole("textbox").fill(snapshot?.shortId ?? "");
    await dialog.getByRole("button", { name: "Delete" }).click();
    await expect(row).toHaveCount(0);

    await repoAction(page, repo, "Repo Sync");
    expect(await snapshotsIn(api, repo)).toEqual([]);
});

// retentionShown is the repository's retention as its form shows it, its edit
// form left open.
async function retentionShown(page: Page, repo: string): Promise<string[]> {
    await page.goto("/integrations/backup-repos/");
    await page
        .getByRole("row", { name: new RegExp(repo) })
        .getByRole("button", { name: "Edit Backup Repo" })
        .click();
    const rules = ["Keep Last", "Keep Hourly", "Keep Daily", "Keep Weekly", "Keep Monthly"];
    const values: string[] = [];
    for (const rule of rules) {
        values.push(await page.getByRole("group", { name: rule, exact: true }).getByRole("spinbutton").inputValue());
    }
    return values;
}

// A repository's retention is what it is created with, and what it is changed
// to: the repository itself is told, so a sync reads it back unchanged, and a
// backup keeps that many snapshots.
test("a repository keeps the snapshots its retention says, as created and as changed", async ({
    page,
    api,
    cleanup,
}) => {
    const store = await storeVolume(api, cleanup, "retains");
    await page.goto("/integrations/backup-repos/create/");
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill(store.name);
    await page.getByRole("group", { name: "Type" }).getByRole("tab", { name: "Volume" }).click();
    await page.getByRole("group", { name: "Volume", exact: true }).getByRole("combobox").click();
    await page.getByRole("option", { name: store.name }).click();
    await page.getByRole("group", { name: "Password *" }).getByRole("textbox").fill(REPO_PASSWORD);
    await page.getByRole("group", { name: "Available in Projects" }).getByRole("checkbox").check();
    const keep = (rule: string) => page.getByRole("group", { name: rule, exact: true }).getByRole("spinbutton");
    await keep("Keep Last").fill("2");
    for (const rule of ["Keep Hourly", "Keep Daily", "Keep Weekly", "Keep Monthly"]) {
        await keep(rule).fill("0");
    }
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: new RegExp(store.name) })).toBeVisible();
    const repo = { id: await settingIdNamed(api, "settings/backup-repos", store.name), name: store.name };
    expect(await retentionShown(page, repo.name)).toEqual(["2", "0", "0", "0", "0"]);

    // A backup expires what the retention does not keep, as it is made.
    const { app, volumeId } = await dataApp(page, api, cleanup, "retain", "echo kept > /data/f");
    const job = await volumeBackupJob(api, app, "e2e-backup", volumeId, repo);
    for (let i = 0; i < 3; i++) {
        await backedUp(api, app, job);
    }
    expect(await snapshotsIn(api, repo)).toHaveLength(2);

    await retentionShown(page, repo.name);
    await keep("Keep Last").fill("1");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: new RegExp(repo.name) })).toBeVisible();
    await repoAction(page, repo, "Repo Sync");
    expect(await retentionShown(page, repo.name)).toEqual(["1", "0", "0", "0", "0"]);
    await backedUp(api, app, job);
    expect(await snapshotsIn(api, repo)).toHaveLength(1);
});

// A command backup whose command fails backs up nothing: a part of a dump is
// not a snapshot to restore from.
test("a command backup that fails leaves no snapshot", async ({ page, api, cleanup }) => {
    const repo = await volumeRepo(api, cleanup, "fails");
    const app = await appIn(api, cleanup, "failing-dump");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");

    const job = await created(api, `${appPath(app)}/sched-jobs`, {
        name: "e2e-dump",
        jobType: "data-backup",
        app: { id: app.id },
        dataBackup: {
            source: "command",
            sourceCommand: { command: "sh -c 'echo partial; exit 3'" },
            sourceFileName: "db.sql",
            restoreCommand: { command: "cat" },
            targetRepository: { id: repo.id },
        },
    });
    const result = await runJob(api, app, job);
    expect(result.status, result.log).toBe("failed");
    expect(await snapshotsIn(api, repo)).toEqual([]);
});
