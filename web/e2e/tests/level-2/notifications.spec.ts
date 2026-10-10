import type { APIRequestContext, Page } from "@playwright/test";

import {
    type App,
    appPath,
    createApp,
    createProject,
    deleteProject,
    deployImage,
    exposeApp,
    latestDeployment,
} from "../../support/api";
import { BUSYBOX, DEPLOYED, WHOAMI, appPage, copyShownLogs, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { expectRuns, newHealthCheck, setHealth } from "../../support/health-checks";
import { echoHook, slackField as field, slackMessages, slackTarget } from "../../support/notifications";
import { domainFor, visit } from "../../support/routing";

test.describe.configure({ timeout: 360_000 });

// told is what the hook was sent about a health check, in the order it came:
// "failed" or "succeeded" for each message naming the check.
async function told(page: Page, hook: App, check: string): Promise<string[]> {
    // The line the server starts with: the log is there once it is shown.
    await expectLogs(page, hook, "Listening on ports");
    return (await copyShownLogs(page))
        .split("\n")
        .filter(line => line.includes(`"value":"${check}"`))
        .map(line => /Healthcheck (failed|succeeded)/.exec(line)?.[1] ?? line);
}

test("a failing health check is told to Slack once, and again when it passes; Repeat while failing tells it again", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("notifications"));
    cleanup(() => deleteProject(api, project.id));
    const { hook, domain: hookDomain } = await echoHook(page, api, project, "notifications-hook");
    const web = await createApp(api, project, "web");
    await deployImage(api, web, WHOAMI);
    await deployed(api, web);
    // Over plain HTTP, as the hook: the backend calls it.
    const webDomain = domainFor("notifications-web");
    await exposeApp(api, web, webDomain, { forceHttps: false });
    await visit(page, `http://${webDomain}/health`);

    const field = (name: string) => page.getByRole("group", { name, exact: true });

    // The project's Slack, tried before it is saved.
    const platform = e2eName("slack");
    await page.goto(`/projects/${project.id}/integrations/im-platforms/create/`);
    await field("Name").getByRole("textbox").fill(platform);
    await expect(field("Type").getByRole("combobox")).toHaveText("Slack Webhook");
    await field("Webhook URL *").getByRole("textbox").fill(`http://${hookDomain}/slack`);
    await page.getByRole("button", { name: "Test Send Msg" }).click();
    await expect(page.getByText("Succeeded", { exact: true })).toBeVisible();
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Project IM platform created successfully")).toBeVisible();
    await expectLogs(page, hook, '"text":"test message"');

    // A target that sends through it.
    const target = e2eName("to-slack");
    await page.goto(`/projects/${project.id}/integrations/notification-targets/create/`);
    await field("Name").getByRole("textbox").fill(target);
    const slack = page.getByRole("region", { name: "Slack Notification" });
    await slack.getByRole("group", { name: "Enabled" }).getByRole("checkbox").check();
    await slack.getByRole("group", { name: "Use Default Slack Webhook" }).getByRole("checkbox").uncheck();
    // The form picks the project's Slack when it is the only one it sees; it is
    // chosen otherwise - choosing it once picked would clear it.
    const webhook = slack.getByRole("group", { name: "Webhook", exact: true }).getByRole("combobox");
    await webhook.click();
    const option = page.getByRole("option", { name: platform });
    await expect(option).toBeVisible();
    if ((await webhook.textContent()) === platform) {
        await page.keyboard.press("Escape");
    } else {
        await option.click();
    }
    await expect(webhook).toHaveText(platform);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Project notification target created successfully")).toBeVisible();

    // Two checks of the app's health telling the target, one of them again
    // every 20 seconds while it fails.
    const tellTarget = (repeat?: string) => async () => {
        for (const when of ["On Failure", "On Success"]) {
            await field(`${when} Use Default`).getByRole("checkbox").uncheck();
            await field(when).getByRole("combobox").click();
            await page.getByRole("option", { name: target }).click();
            await expect(field(when).getByRole("combobox")).toHaveText(target);
            await expect(page.getByRole("option")).toHaveCount(0);
        }
        if (repeat) {
            await field("Repeat while failing").getByRole("textbox").fill(repeat);
        }
    };
    const health = `http://${webDomain}/health`;
    await newHealthCheck(page, web, "check-once", health, tellTarget());
    await newHealthCheck(page, web, "check-again", health, tellTarget("20s"));

    // Healthy from their first run, they tell nothing.
    await expectRuns(page, api, web, "check-once", "Done", "Failed");
    await expectRuns(page, api, web, "check-again", "Done", "Failed");
    expect(await told(page, hook, "check-once")).toEqual([]);
    expect(await told(page, hook, "check-again")).toEqual([]);

    // Failing: each tells it, and only the one asked to tells it again.
    await setHealth(page, webDomain, 500);
    await expect
        .poll(async () => (await told(page, hook, "check-again")).length, { timeout: 120_000, intervals: [5_000] })
        .toBeGreaterThanOrEqual(2);
    expect(new Set(await told(page, hook, "check-again"))).toEqual(new Set(["failed"]));
    expect(await told(page, hook, "check-once")).toEqual(["failed"]);

    // Healthy again: each tells it once, and nothing more.
    await setHealth(page, webDomain, 200);
    await expect
        .poll(async () => told(page, hook, "check-once"), { timeout: 60_000, intervals: [5_000] })
        .toEqual(["failed", "succeeded"]);
    await expect
        .poll(async () => (await told(page, hook, "check-again")).at(-1), { timeout: 60_000, intervals: [5_000] })
        .toBe("succeeded");
    const again = await told(page, hook, "check-again");
    await page.waitForTimeout(25_000);
    expect(await told(page, hook, "check-once")).toEqual(["failed", "succeeded"]);
    expect(await told(page, hook, "check-again")).toEqual(again);
});

// deployTelling deploys an image to the app, telling the target how its
// deployment ends - done or failed - and nobody else.
async function deployTelling(api: APIRequestContext, app: App, image: string, target: string): Promise<string> {
    const current = await api.get(`${appPath(app)}/deployment-settings`);
    const { updateVer } = ((await current.json()) as { data: { updateVer: number } }).data;
    const res = await api.put(`${appPath(app)}/deployment-settings`, {
        data: {
            entrypoint: "",
            command: "",
            workingDir: "",
            preDeploymentCommand: "",
            postDeploymentCommand: "",
            notification: {
                success: { id: target },
                successUseDefault: false,
                failure: { id: target },
                failureUseDefault: false,
            },
            activeMethod: "image",
            imageSource: { image, registryAuth: { id: "" } },
            updateVer,
        },
    });
    expect(res.ok(), `deploying ${image}: ${res.status()} ${await res.text()}`).toBe(true);
    return ((await res.json()) as { data: { deploymentId: string } }).data.deploymentId;
}

// A deployment tells the app's target how it ended: failed, and why, or done -
// the image, and where its details are.
test("a deployment tells its notification target that it failed, and why, and that it is done", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("deploy-told"));
    cleanup(() => deleteProject(api, project.id));
    const { hook, id: target } = await slackTarget(page, api, project, "deploy-told-hook");
    const ofWeb = async () => (await slackMessages(page, hook)).filter(m => m.title.includes("[web]"));

    const web = await createApp(api, project, "web");
    const failedId = await deployTelling(api, web, "traefik/whoami:e2e-no-such-tag", target);
    await expect.poll(async () => (await latestDeployment(api, web))?.status, DEPLOYED).toBe("failed");
    await expect
        .poll(async () => (await ofWeb()).map(m => m.title), { timeout: 60_000, intervals: [5_000] })
        .toEqual([`[${project.name}][web] Deployment failed`]);
    const [failed] = await ofWeb();
    expect(field(failed, "Image")).toBe("traefik/whoami:e2e-no-such-tag");
    // Why, as the deployment's details say it.
    expect(field(failed, "Reason")).toContain("whoami:e2e-no-such-tag: not found");
    expect(field(failed, "See deployment details")).toContain(`/deployments/${failedId}|`);

    await deployTelling(api, web, WHOAMI, target);
    await deployed(api, web);
    await expect.poll(async () => (await ofWeb()).length, { timeout: 60_000, intervals: [5_000] }).toBe(2);
    const done = (await ofWeb())[1];
    expect(done?.title).toBe(`[${project.name}][web] Deployment succeeded`);
    expect(field(done, "Image")).toBe(WHOAMI);
    expect(field(done, "Reason"), "nothing failed").toBeUndefined();
});

// A scheduled job tells the target its form names when a run fails: the job,
// what it failed with.
test("a scheduled job that fails tells its notification target", async ({ page, api, cleanup }) => {
    const project = await createProject(api, e2eName("job-told"));
    cleanup(() => deleteProject(api, project.id));
    const { hook, name: target } = await slackTarget(page, api, project, "job-told-hook");

    const web = await createApp(api, project, "web");
    await deployImage(api, web, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, web);
    await expectLogs(page, web, "ready");
    await page.goto(appPage(web, "sched-jobs"));
    await page.getByRole("button", { name: "New Scheduled Job" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-breaks");
    await page.getByRole("group", { name: "Scheduling Mode" }).getByRole("tab", { name: "No schedule" }).click();
    await page.getByRole("group", { name: "Command *" }).getByRole("textbox").fill("sh -c 'echo broken; exit 3'");
    await page.getByRole("checkbox", { name: "On Failure Use Default" }).uncheck();
    await page.getByRole("group", { name: "On Failure", exact: true }).getByRole("combobox").click();
    await page.getByRole("option", { name: target }).click();
    await page.getByRole("button", { name: "Save" }).click();
    const row = page.getByRole("row", { name: /e2e-breaks/ });
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: "Run Now" }).click();

    const ofJob = async () => (await slackMessages(page, hook)).filter(m => field(m, "Scheduled Job") === "e2e-breaks");
    await expect
        .poll(async () => (await ofJob()).map(m => m.title), { timeout: 90_000, intervals: [5_000] })
        .toEqual([`[${project.name}][web] Scheduled task failed`]);
});
