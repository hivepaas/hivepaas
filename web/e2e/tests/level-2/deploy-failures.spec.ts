import type { APIRequestContext } from "@playwright/test";

import { type App, appPath, deployImage, latestDeployment } from "../../support/api";
import {
    BUSYBOX,
    DEPLOYED,
    WHOAMI,
    appIn,
    appPage,
    deployed,
    expectInstances,
    expectShownLogs,
} from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

interface ServiceTask {
    id: string;
    desiredState: string;
    status: { state: string };
}

// runningTask is the app's container once swarm has settled on it: the only
// one meant to run, and running.
async function runningTask(api: APIRequestContext, app: App): Promise<string> {
    let id = "";
    await expect
        .poll(
            async () => {
                const body = (await (await api.get(`${appPath(app)}/service-tasks`)).json()) as {
                    data: ServiceTask[];
                };
                const meant = body.data.filter(task => task.desiredState === "running");
                id = meant.length === 1 && meant[0]?.status.state === "running" ? meant[0].id : "";
                return id;
            },
            { timeout: 60_000 },
        )
        .not.toBe("");
    return id;
}

test("a deployment of an image that does not exist fails, says why, and leaves the app running", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "no-image");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    const running = await runningTask(api, app);

    const id = await deployImage(api, app, "traefik/whoami:e2e-no-such-tag");
    await expect.poll(async () => (await latestDeployment(api, app))?.status, DEPLOYED).toBe("failed");

    await page.goto(appPage(app, "deployments"));
    // Why, as a person reads it: the error's code and what it means - here
    // docker's own words - not the chain of codes behind it.
    const failed = page.getByRole("button", { name: `Failed Deployment #${id.slice(0, 8)}` });
    await expect(failed).toContainText("whoami:e2e-no-such-tag: not found");
    await expect(failed).not.toContainText("ERR_NOT_FOUND");
    // The image is pulled before the service is touched: the container the app
    // ran is the one it runs.
    expect(await runningTask(api, app)).toBe(running);
    await expectInstances(page, app, "1/1");
});

test("a container that keeps exiting is on Home with its exit code, and View logs shows what it printed", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "crash-loop");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo e2e-crash-line; exit 3'");

    // Swarm starts it again each time it exits: a few times, and it is told,
    // with the latest failure. In a full run the first tasks can fail before
    // the container runs - "Pool overlaps with other one": the env's new network
    // was given a subnet the node had not let go of yet.
    const item = page
        .getByRole("listitem")
        .filter({ hasText: "web keeps restarting" })
        .filter({ hasText: e2eName("crash-loop") });
    await expect(async () => {
        await page.goto("/home/");
        await expect(item).toContainText("task: non-zero exit (3)", { timeout: 3_000 });
    }).toPass({ timeout: 120_000, intervals: [5_000] });
    // Its env by the name the app's screens use, and the link to them with it.
    await expect(item).toContainText(`${e2eName("crash-loop")} / development`);

    await item.getByRole("link", { name: "View logs" }).click();
    await expect(page).toHaveURL(new RegExp(`${appPage(app, "logs")}$`));
    await expectShownLogs(page, "e2e-crash-line");
});
