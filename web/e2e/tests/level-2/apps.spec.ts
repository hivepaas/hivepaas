import { deployImage, findApp } from "../../support/api";
import { BUSYBOX, WHOAMI, appIn, appPage, deployed, expectInstances, expectLogs, redeploy } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

// Deploying pulls an image and starts its containers: each test has room.
test.describe.configure({ timeout: 240_000 });

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

test("a runtime variable reaches the container once redeployed, as its log shows", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "env-var");
    await deployImage(api, app, BUSYBOX, 'sh -c "echo hello-$E2E_GREETING && exec sleep 3600"');
    await deployed(api, app);
    const greeting = e2eName("greeting");

    await page.goto(appPage(app, "env-variables"));
    const runtime = page.getByRole("region", { name: "Runtime Env Variables", exact: true });
    await runtime.getByRole("button", { name: "Add" }).click();
    await runtime
        .getByRole("group", { name: "New variable" })
        .getByRole("textbox", { name: "Key" })
        .fill("E2E_GREETING");
    await runtime
        .getByRole("group", { name: "E2E_GREETING", exact: true })
        .getByRole("textbox", { name: "Value" })
        .fill(greeting);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();

    await redeploy(page, api, app);

    await expectLogs(page, app, `hello-${greeting}`);
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
