import type { APIRequestContext, Page } from "@playwright/test";

import {
    type App,
    type Project,
    appPath,
    createProject,
    createVolume,
    deleteProject,
    deleteVolume,
    findApp,
    findAppNamed,
} from "../../support/api";
import { deployed, expectInstances } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { domainFor, visit } from "../../support/routing";

test.describe.configure({ timeout: 300_000 });

// The proxy's certificate for a .localhost name is its own, signed by no one.
test.use({ ignoreHTTPSErrors: true });

// The catalog is the app-templates checkout env/up.sh gives the backend.
test("an app made from a template is deployed, and runs", async ({ page, api, cleanup }) => {
    const project = await createProject(api, e2eName("template"));
    cleanup(() => deleteProject(api, project.id));

    await page.goto(`/projects/${project.id}/app-templates/`);
    await page.getByRole("textbox", { name: /^Search templates/ }).fill("IT Tools");
    await page
        .getByRole("group", { name: "IT Tools", exact: true })
        .getByRole("button", { name: "Deploy", exact: true })
        .click();
    const dialog = page.getByRole("dialog", { name: "Deploy Template: IT Tools" });
    await expect(dialog.getByRole("textbox", { name: "e.g. my-app" })).toHaveValue("it-tools");
    await dialog.getByRole("button", { name: "Deploy Template" }).click();

    const near: App = { id: "", name: "", projectId: project.id, env: "development" };
    await expect.poll(() => findAppNamed(api, near, "it-tools"), { timeout: 60_000 }).toBeDefined();
    const app = (await findAppNamed(api, near, "it-tools"))!;
    await deployed(api, app);
    // Its health check passes: the page is served.
    await expectInstances(page, app, "1/1");
});

// companionNamed is the app of that name created with the app - a template's
// dependency - which belongs to it, and is listed under it, not beside it.
async function companionNamed(api: APIRequestContext, owner: App, name: string): Promise<App | undefined> {
    const res = await api.get(`projects/${owner.projectId}/${owner.env}/apps`, { params: { getChildApps: true } });
    expect(res.ok(), `listing apps: ${res.status()}`).toBe(true);
    const apps = ((await res.json()) as { data: { id: string; logicalChildApps?: { id: string; name: string }[] }[] })
        .data;
    const found = apps.find(a => a.id === owner.id)?.logicalChildApps?.find(a => a.name === name);
    return found ? { id: found.id, name, projectId: owner.projectId, env: owner.env } : undefined;
}

// Atuin's sync server, which comes with its database: a template that depends
// on another - postgres - deploys both, and wires one to the other.
test("a template that depends on a database deploys both, the app connected to it", async ({ page, api, cleanup }) => {
    const project = await createProject(api, e2eName("template-deps"));
    cleanup(() => deleteProject(api, project.id));
    const domain = domainFor("atuin");

    await page.goto(`/projects/${project.id}/app-templates/`);
    await page.getByRole("textbox", { name: /^Search templates/ }).fill("Atuin");
    await page
        .getByRole("group", { name: "Atuin", exact: true })
        .getByRole("button", { name: "Deploy", exact: true })
        .click();
    const dialog = page.getByRole("dialog", { name: "Deploy Template: Atuin" });
    await expect(dialog.getByRole("textbox", { name: "e.g. my-app" })).toHaveValue("atuin");
    // What comes with it, before anything is made.
    await expect(dialog.getByText("Dependent Services")).toBeVisible();
    await expect(dialog.getByText("Database (postgres)")).toBeVisible();
    await dialog.getByRole("textbox", { name: "e.g. app.example.com" }).fill(domain);
    await dialog.getByRole("button", { name: "Deploy Template" }).click();

    const near: App = { id: "", name: "", projectId: project.id, env: "development" };
    await expect.poll(() => findAppNamed(api, near, "atuin"), { timeout: 60_000 }).toBeDefined();
    const atuin = (await findAppNamed(api, near, "atuin"))!;
    const db = await companionNamed(api, atuin, "atuin-db");
    expect(db, "its database is an app of its own, named after it").toBeDefined();
    await deployed(api, db!);
    await deployed(api, atuin);

    const res = await visit(page, `https://${domain}/`);
    expect(((await res.json()) as { version?: string }).version, "Atuin answers at its domain").toBeTruthy();
    // An account made is kept in the database, and read back from it.
    const account = await page.evaluate(async () => {
        const made = await fetch("/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: "e2e", email: "e2e@example.com", password: "e2e-password" }),
        });
        const read = await fetch("/user/e2e");
        return { made: made.status, read: read.status, user: await read.text() };
    });
    expect(account.made, "an account is made").toBe(200);
    expect(account.read, "and read back").toBe(200);
    expect(account.user).toContain('"username":"e2e"');
});

// openDeploy opens a template's Deploy dialog from the store, the template
// found by searching for its title.
async function openDeploy(page: Page, project: Project, title: string) {
    await page.goto(`/projects/${project.id}/app-templates/`);
    await page.getByRole("textbox", { name: /^Search templates/ }).fill(title);
    await page
        .getByRole("group", { name: title, exact: true })
        .getByRole("button", { name: "Deploy", exact: true })
        .click();
    return page.getByRole("dialog", { name: `Deploy Template: ${title}` });
}

// madeAnew is the app of that name once there is one, other than the one
// before it.
async function madeAnew(api: APIRequestContext, near: App, name: string, before?: App): Promise<App> {
    let app: App | undefined;
    await expect
        .poll(
            async () => {
                app = await findAppNamed(api, near, name);
                return app !== undefined && app.id !== before?.id;
            },
            { message: `${name} is made`, timeout: 60_000 },
        )
        .toBe(true);
    return app!;
}

// deleteKeepingData deletes an app, and leaves what it stored where it is.
async function deleteKeepingData(api: APIRequestContext, app: App): Promise<void> {
    const res = await api.delete(appPath(app));
    expect(res.ok(), `deleting ${app.name}: ${res.status()} ${await res.text()}`).toBe(true);
    await expect.poll(() => findApp(api, app), { timeout: 60_000 }).toBeUndefined();
}

// deployNtfy deploys ntfy from the store, its data on the volume.
async function deployNtfy(page: Page, project: Project, volume: string): Promise<void> {
    const dialog = await openDeploy(page, project, "ntfy");
    await dialog.getByRole("group", { name: "Data volume" }).getByRole("combobox").click();
    await page.getByRole("option", { name: volume }).click();
    await dialog.getByRole("button", { name: "Deploy Template" }).click();
}

const DATA_DIR = "/var/lib/ntfy";
const MARKER = { name: "e2e-marker.txt", mimeType: "text/plain", buffer: Buffer.from("left by the first install") };

// download is a file of the app's container, once it runs one.
async function download(api: APIRequestContext, app: App, path: string, isDir = false) {
    return api.get(`${appPath(app)}/container/file-download`, { params: { path, isDir } });
}

// ntfy keeps its data on the volume it is given, in a directory of the app's -
// named by its key. Deleting the app leaves the directory; an app made again by that name
// would start with it. Before it is made, the dialog says so, and asks: start
// with that data, or delete it first.
test("a template deployed where a deleted app left its data asks first, and keeps it or deletes it", async ({
    page,
    api,
    cleanup,
}) => {
    const volume = e2eName("ntfy-data");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    const project = await createProject(api, e2eName("template-data"));
    cleanup(() => deleteProject(api, project.id));
    const near: App = { id: "", name: "", projectId: project.id, env: "development" };

    await deployNtfy(page, project, volume);
    let app = await madeAnew(api, near, "ntfy");
    await deployed(api, app);
    // A file of its own in its data, put there once its container runs.
    await expect(async () => {
        const res = await api.post(`${appPath(app)}/container/file-upload`, {
            multipart: { file: MARKER, path: `${DATA_DIR}/` },
        });
        expect(res.ok(), `uploading the marker: ${res.status()} ${await res.text()}`).toBe(true);
    }).toPass({ timeout: 60_000, intervals: [3_000] });
    await deleteKeepingData(api, app);

    // Made again, on what it left: kept.
    await deployNtfy(page, project, volume);
    let inUse = page.getByRole("dialog", { name: "These apps already have data" });
    await expect(inUse).toContainText("One app would be created on a directory that a previous install left behind");
    await expect(inUse.getByRole("listitem")).toContainText("ntfy");
    await inUse.getByRole("button", { name: "Create anyway" }).click();
    let before = app;
    app = await madeAnew(api, near, "ntfy", before);
    await deployed(api, app);
    await expect(async () => {
        const res = await download(api, app, `${DATA_DIR}/${MARKER.name}`);
        expect(res.status()).toBe(200);
        expect(await res.text()).toBe(MARKER.buffer.toString());
    }, "the data left behind is the new app's").toPass({ timeout: 60_000, intervals: [3_000] });
    await deleteKeepingData(api, app);

    // Made again, the data deleted first: it starts without it.
    await deployNtfy(page, project, volume);
    inUse = page.getByRole("dialog", { name: "These apps already have data" });
    await inUse.getByRole("button", { name: "Delete that data and create" }).click();
    before = app;
    app = await madeAnew(api, near, "ntfy", before);
    await deployed(api, app);
    await expect(async () => {
        expect((await download(api, app, DATA_DIR, true)).status(), "its container runs").toBe(200);
    }).toPass({ timeout: 60_000, intervals: [3_000] });
    expect((await download(api, app, `${DATA_DIR}/${MARKER.name}`)).status(), "the data is gone").not.toBe(200);
});
