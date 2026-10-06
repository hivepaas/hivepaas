import type { APIRequestContext, Page } from "@playwright/test";

import {
    type App,
    createApp,
    createProject,
    deleteProject,
    deployImage,
    findApp,
    latestDeployment,
} from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";

// Deploying pulls an image and starts its containers: each test has room.
test.describe.configure({ timeout: 240_000 });

const WHOAMI = "traefik/whoami:latest";
const BUSYBOX = "busybox:1.37";
const DEPLOYED = { timeout: 180_000, intervals: [2_000] };

// appIn makes a project of its own with one app in it; the project, and the
// app with it, go once the test ends.
async function appIn(api: APIRequestContext, cleanup: (step: () => Promise<unknown>) => void, label: string) {
    const project = await createProject(api, e2eName(label));
    cleanup(() => deleteProject(api, project.id));
    return createApp(api, project, "web");
}

// deployed waits for the app's deployment - its newest, or the one given - to
// be done.
async function deployed(api: APIRequestContext, app: App, id?: string): Promise<void> {
    await expect
        .poll(async () => {
            const latest = await latestDeployment(api, app);
            return id && latest?.id !== id ? "not yet" : latest?.status;
        }, DEPLOYED)
        .toBe("done");
}

// instances reads, after a reload, how many of the app's replicas run against
// how many are asked for, as the app's header says it: "1/1".
async function expectInstances(page: Page, app: App, value: string): Promise<void> {
    await expect(async () => {
        await page.goto(`/projects/${app.projectId}/${app.env}/apps/${app.id}/general/`);
        await expect(page.getByRole("link", { name: /^Instances/ })).toContainText(value, { timeout: 2_000 });
    }).toPass({ timeout: 120_000, intervals: [3_000] });
}

const appPage = (app: App, tab: string) => `/projects/${app.projectId}/${app.env}/apps/${app.id}/${tab}/`;

test("an app deployed from an image runs, and its deployment is listed as done", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "deploy");
    await page.goto(appPage(app, "deployment-settings"));

    await page
        .getByRole("group", { name: /^Docker Image/ })
        .getByRole("textbox")
        .fill(WHOAMI);
    await page.getByRole("button", { name: "Deploy", exact: true }).click();

    await deployed(api, app);
    await expectInstances(page, app, "1/1");
    await page.goto(appPage(app, "deployments"));
    await expect(page.getByRole("button", { name: new RegExp(`^Done ${WHOAMI}`) })).toBeVisible();
});

test("a runtime variable reaches the container once redeployed, as its log shows", async ({
    page,
    api,
    cleanup,
    context,
}) => {
    await context.grantPermissions(["clipboard-read", "clipboard-write"]);
    const app = await appIn(api, cleanup, "env-var");
    await deployImage(api, app, BUSYBOX, 'sh -c "echo hello-$E2E_GREETING && exec sleep 3600"');
    await deployed(api, app);
    const first = await latestDeployment(api, app);
    const greeting = e2eName("greeting");

    await page.goto(appPage(app, "env-variables"));
    const runtime = page.getByRole("region", { name: "Runtime Env Variables", exact: true });
    await runtime.getByRole("button", { name: "Add" }).click();
    await runtime.getByRole("textbox", { name: "Key" }).last().fill("E2E_GREETING");
    // The value's field has no name of its own: the one after the key.
    await runtime.getByRole("textbox").last().fill(greeting);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();

    await page.getByRole("button", { name: "Re-deploy" }).click();
    await page.getByRole("dialog", { name: "Re-deploy app" }).getByRole("button", { name: "Re-deploy" }).click();
    await expect.poll(async () => (await latestDeployment(api, app))?.id, DEPLOYED).not.toBe(first?.id);
    await deployed(api, app);

    await page.goto(appPage(app, "logs"));
    await expect(async () => {
        await page.getByRole("button", { name: "Copy logs" }).locator("visible=true").first().click();
        const copied = await page.evaluate(() => navigator.clipboard.readText());
        expect(copied).toContain(`hello-${greeting}`);
    }).toPass({ timeout: 60_000, intervals: [2_000] });
});

test("scaled to two, an app runs two replicas", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "scale");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    await page.goto(appPage(app, "availability-and-scaling"));

    await page.getByRole("group", { name: "Replicas" }).getByRole("textbox").fill("2");
    await page.getByRole("button", { name: "Save", exact: true }).click();

    await expectInstances(page, app, "2/2");
});

test("a stopped app runs no replica, and runs again once started", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "stop");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    await expectInstances(page, app, "1/1");

    // Names matched whole: "Start" is in "Restart".
    await page.getByRole("button", { name: "Stop", exact: true }).click();
    await page.getByRole("dialog", { name: "Stop app" }).getByRole("button", { name: "Stop", exact: true }).click();
    await expectInstances(page, app, "0/0");

    await page.getByRole("button", { name: "Start", exact: true }).click();
    await expectInstances(page, app, "1/1");
});

test("deleting an app asks for its name, then removes it", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "delete");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    await page.goto(appPage(app, "danger-zone"));

    await page.getByRole("button", { name: "Delete App" }).click();
    const dialog = page.getByRole("dialog", { name: "Delete app" });
    const confirm = dialog.getByRole("button", { name: "Delete this App" });
    await expect(confirm).toBeDisabled();
    await dialog.getByRole("textbox").fill(app.name);
    await dialog.getByRole("checkbox", { name: /Also delete the stored data/ }).check();
    await confirm.click();

    await expect.poll(() => findApp(api, app)).toBeUndefined();
});
