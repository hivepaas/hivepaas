import {
    createJob,
    createVolume,
    deleteSettingsNamed,
    deleteVolume,
    deployImage,
    runJob,
    setRuntimeEnvVars,
} from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs, mountVolume } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

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
