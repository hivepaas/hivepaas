import type { Page } from "@playwright/test";

import { deployImage, latestDeployment } from "../../support/api";
import { BUSYBOX, DEPLOYED, WHOAMI, appIn, appPage, deployed, expectShownLogs, runningTask } from "../../support/apps";
import { expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

// cards are the deployments' cards on the app's Deployments tab, the newest
// first, each named by its status, then what it deployed.
const cards = (page: Page) => page.getByRole("button", { name: /^(Done|Failed|In-Progress|Not Started|Canceled) / });

// A deployment runs its pre-deployment command in the app's running container
// before it touches the service: one that sleeps holds it in progress, for as
// long as it takes to cancel it.
test("a deployment in progress is canceled from its card, and the app runs on as it ran", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "cancel");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo first; exec sleep 3600'");
    await deployed(api, app);
    const running = await runningTask(api, app);

    await deployImage(api, app, BUSYBOX, "sh -c 'echo second; exec sleep 3600'", "sleep 600");
    await page.goto(appPage(app, "deployments"));
    const latest = cards(page).first();
    await expect(latest).toContainText("In-Progress", { timeout: 60_000 });
    await latest.getByRole("button", { name: "Cancel", exact: true }).click();
    await page.getByRole("button", { name: "Cancel deployment", exact: true }).click();

    await expect.poll(async () => (await latestDeployment(api, app))?.status, DEPLOYED).toBe("canceled");
    await expect(latest).toContainText("Canceled");
    // Canceled before the service was touched: the container it ran is the one
    // it runs.
    expect(await runningTask(api, app)).toBe(running);
});

// Each deployment is listed, the newest first, and keeps its own log: what the
// second's pre-deployment command printed is in the second's alone. The second
// comes right after the first: its command waits for the container the first
// started to have run long enough - 15 seconds - and runs in it.
test("each deployment is listed, the newest first, and its log is its own", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "history");
    const first = await deployImage(api, app, BUSYBOX, "sh -c 'exec sleep 3600'");
    await deployed(api, app, first);
    const second = await deployImage(api, app, WHOAMI, "", "echo ran-before-second");
    await deployed(api, app, second);

    await page.goto(appPage(app, "deployments"));
    await expect(cards(page)).toHaveCount(2);
    await expect(cards(page).nth(0)).toContainText("traefik/whoami");
    await expect(cards(page).nth(1)).toContainText("busybox:1.37");

    // The older one, opened from its card.
    await cards(page).nth(1).click();
    await expect(page).toHaveURL(new RegExp(`/deployments/${first}/$`));
    const older = await expectShownLogs(page, /Deployment finished[\s\S]*/);
    expect(older, "the first deployment's log").not.toContain("ran-before-second");

    await page.goto(appPage(app, `deployments/${second}`));
    await expectShownLogs(page, "ran-before-second");
});
