import { deployImage, findAppNamed, setRuntimeEnvVars } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs, redeploy } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

test("an app cloned in its environment runs, with its image and variables, and deploys", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "clone");
    const greeting = e2eName("greeting");
    await setRuntimeEnvVars(api, app, [{ key: "E2E_GREETING", value: greeting, isLiteral: false }]);
    await deployImage(api, app, BUSYBOX, "sh -c 'echo \"hello-$E2E_GREETING\"; exec sleep 3600'");
    await deployed(api, app);

    await page.goto(appPage(app, "app-clone"));
    await page.getByRole("group", { name: "Target Name" }).getByRole("textbox").fill("web-clone");
    await page.getByRole("button", { name: "Save and Clone" }).click();
    await expect(page.getByText("App clone started")).toBeVisible();

    // The clone goes with the project, as the app does.
    await expect.poll(() => findAppNamed(api, app, "web-clone"), { timeout: 120_000 }).toBeDefined();
    const clone = (await findAppNamed(api, app, "web-clone"))!;
    // A clone is made from the app's service, not deployed: it lists no
    // deployment, and its log is what says it runs.
    await expectLogs(page, clone, `hello-${greeting}`);

    // Its deployment settings came along: it deploys on its own.
    await redeploy(page, api, clone);
});
