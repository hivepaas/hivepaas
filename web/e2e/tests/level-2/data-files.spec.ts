import { type App, appPath, createJob, deployImage, runJob, setRuntimeEnvVars } from "../../support/api";
import { appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

// The database a job dumps, and a data file is loaded back into.
const POSTGRES = "postgres:18.6-alpine";
const PASSPHRASE = "E2e-dump-passphrase-1";

test.describe.configure({ timeout: 300_000 });

const psql = (sql: string) => `psql -U postgres -v ON_ERROR_STOP=1 -tAc "${sql}"`;

// ran runs a command in the app as a job of its own, and answers what it
// printed; it is to succeed.
async function ran(api: Parameters<typeof runJob>[0], app: App, name: string, command: string): Promise<string> {
    const job = await createJob(api, app, name, command);
    const result = await runJob(api, app, job);
    expect(result.status, result.log).toBe("done");
    return result.log;
}

// A dump a job saved as a data file - compressed and encrypted, as the job
// was told - is loaded back into the database from Data Files: the file fed
// to the command given, on its stdin, decrypted with its passphrase and
// decompressed. With another passphrase the load fails, and loads nothing.
test("a dump a job saved as a data file is loaded back into the database through a command", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "load-dump");
    await setRuntimeEnvVars(api, app, [{ key: "POSTGRES_PASSWORD", value: "E2e-pg-Pass1", isLiteral: true }]);
    await deployImage(api, app, POSTGRES);
    await deployed(api, app);
    await expectLogs(page, app, "PostgreSQL init process complete");

    const mark = e2eName("note");
    await ran(api, app, "e2e-seed", psql(`CREATE TABLE notes(v text); INSERT INTO notes VALUES ('${mark}')`));
    const dump = await createJob(api, app, "e2e-dump", "pg_dump -U postgres --clean --if-exists postgres", {
        commandOutput: {
            enabled: true,
            saveToFile: {
                fileName: "notes.sql",
                fileKind: "postgres-backup",
                storage: { id: "" },
                compressionFormat: "gzip",
                encryptionFormat: "age",
                encryptionSecret: PASSPHRASE,
            },
        },
    });
    expect((await runJob(api, app, dump)).status).toBe("done");
    await ran(api, app, "e2e-drop", psql("DROP TABLE notes"));

    // Without its passphrase the load is refused; with another it fails, and
    // the table stays gone.
    const files = (await (await api.get(`${appPath(app)}/data-files`)).json()) as {
        data: { id: string; name: string }[];
    };
    const saved = files.data.find(file => file.name.startsWith("notes.sql"));
    expect(saved?.name, "the dump's file, named as compressed then encrypted").toBe("notes.sql.gz.age");
    const load = `${appPath(app)}/data-files/${saved?.id ?? ""}/load`;
    const command = { command: "psql -U postgres -v ON_ERROR_STOP=1 postgres" };
    const unread = await api.post(load, { data: { command } });
    expect(unread.status(), "an encrypted file without its passphrase").toBe(400);
    const wrong = await api.post(load, { data: { command, passphrase: "not-the-passphrase" } });
    expect(wrong.ok(), `loading with another passphrase: ${await wrong.text()}`).toBe(true);
    const wrongTask = ((await wrong.json()) as { data: { task: { id: string } } }).data.task.id;
    await expect
        .poll(
            async () =>
                ((await (await api.get(`${appPath(app)}/tasks/${wrongTask}`)).json()) as { data: { status: string } })
                    .data.status,
            { timeout: 60_000 },
        )
        .toBe("failed");

    await page.goto(appPage(app, "data-files"));
    await page
        .getByRole("row", { name: /notes\.sql\.gz\.age/ })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: "Load into Command" }).click();
    await expect(page.getByRole("heading", { name: "Load notes.sql.gz.age" })).toBeVisible();
    await page
        .getByRole("group", { name: "Command *" })
        .getByRole("textbox")
        .fill("psql -U postgres -v ON_ERROR_STOP=1 postgres");
    // A password field: no textbox to the accessibility tree.
    await page
        .getByRole("group", { name: /^Passphrase/ })
        .locator("input[type=password]")
        .fill(PASSPHRASE);
    await page.getByRole("group", { name: "App Name" }).getByRole("textbox").fill(app.name);
    await page.getByRole("button", { name: "Load", exact: true }).click();
    await expect(page.locator("main")).toContainText(/Done\s*task:/, { timeout: 120_000 });

    expect(await ran(api, app, "e2e-read", psql("SELECT v FROM notes"))).toContain(mark);
});
