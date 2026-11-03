import { deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 240_000 });

test("a config file mounted into the app is read by its container", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "config-file");
    const content = `greeting=${e2eName("hello")}`;

    await page.goto(appPage(app, "config-files"));
    await page.getByRole("button", { name: "New Config File" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("E2E_CONF");
    await page.getByRole("group", { name: "Value *" }).getByRole("textbox").fill(content);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: /E2E_CONF/ })).toBeVisible();

    await page.goto(appPage(app, "setting-mounts"));
    await page.getByRole("button", { name: "New Setting Mount" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-conf");
    await page.getByRole("combobox", { name: "Mount From" }).click();
    await page.getByRole("option", { name: "Config file" }).click();
    await page.getByRole("group", { name: "Setting *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: "E2E_CONF" }).click();
    await page.getByRole("checkbox", { name: "content", exact: true }).check();
    await page.getByRole("textbox", { name: "content path" }).fill("/etc/e2e.conf");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: /e2e-conf/ })).toBeVisible();

    // The file has no newline of its own, and docker logs a line once it ends.
    await deployImage(api, app, BUSYBOX, "sh -c 'cat /etc/e2e.conf; echo; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, content);
});

test("a secret referenced from a variable reaches the container", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "secret");
    const token = e2eName("token");

    await page.goto(appPage(app, "secrets"));
    await page.getByRole("button", { name: "New Secret" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("E2E_TOKEN");
    await page.getByRole("group", { name: "Value *" }).getByRole("textbox").fill(token);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: /E2E_TOKEN/ })).toBeVisible();

    await page.goto(appPage(app, "env-variables"));
    const runtime = page.getByRole("region", { name: "Runtime Env Variables", exact: true });
    await runtime.getByRole("button", { name: "Add" }).click();
    await runtime
        .getByRole("group", { name: "New variable" })
        .getByRole("textbox", { name: "Key" })
        .fill("E2E_TOKEN_VAR");
    await runtime
        .getByRole("group", { name: "E2E_TOKEN_VAR", exact: true })
        .getByRole("textbox", { name: "Value" })
        .fill("${secrets.E2E_TOKEN}");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();

    await deployImage(api, app, BUSYBOX, "sh -c 'echo \"token-$E2E_TOKEN_VAR\"; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, `token-${token}`);
});
