import type { Page } from "@playwright/test";

import {
    type App,
    createApp,
    createProject,
    deleteProject,
    deployImage,
    exposeApp,
    setRuntimeEnvVars,
} from "../../support/api";
import { WHOAMI, copyShownLogs, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { expectRuns, newHealthCheck } from "../../support/health-checks";
import { domainFor, visit } from "../../support/routing";

test.describe.configure({ timeout: 360_000 });

// What stands in for Slack: an echo server, which logs each request it gets -
// body and all, on a line of its own with LOG_WITHOUT_NEWLINE.
const ECHO = "mendhak/http-https-echo:42";

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

// setHealth has whoami answer its /health with the code from now on: it takes
// the code POSTed there.
async function setHealth(page: Page, domain: string, code: number): Promise<void> {
    await page.goto(`http://${domain}/`);
    await page.evaluate(async body => {
        await fetch("/health", { method: "POST", body });
    }, String(code));
    await visit(page, `http://${domain}/health`, code);
}

test("a failing health check is told to Slack once, and again when it passes; Repeat while failing tells it again", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("notifications"));
    cleanup(() => deleteProject(api, project.id));
    const web = await createApp(api, project, "web");
    const hook = await createApp(api, project, "hook");

    await setRuntimeEnvVars(api, hook, [{ key: "LOG_WITHOUT_NEWLINE", value: "true", isLiteral: true }]);
    await deployImage(api, hook, ECHO);
    await deployImage(api, web, WHOAMI);
    await deployed(api, hook);
    await deployed(api, web);
    // Over plain HTTP: what calls them - the backend - trusts only certificates
    // someone signed, and a .localhost one is signed by no one.
    const hookDomain = domainFor("notifications-hook");
    const webDomain = domainFor("notifications-web");
    await exposeApp(api, hook, hookDomain, { port: 8080, forceHttps: false });
    await exposeApp(api, web, webDomain, { forceHttps: false });
    await visit(page, `http://${hookDomain}/`);
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
