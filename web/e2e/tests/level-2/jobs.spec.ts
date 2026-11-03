import type { APIRequestContext, Page } from "@playwright/test";

import { type App, appPath, createJob, deployImage, latestDeployment } from "../../support/api";
import {
    BUSYBOX,
    appIn,
    appPage,
    copyShownLogs,
    deployed,
    expectLogs,
    expectShownLogs,
    redeploy,
} from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

// How long a run's page takes to say how it ended: it asks every 5 seconds once
// the run's log has ended.
const RUN_ENDED = { timeout: 20_000 };

test("a scheduled job run by hand runs its command in the app's container", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "job");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    // A deployment is done once swarm has the service's new spec; the job needs
    // the container that spec starts.
    await expectLogs(page, app, "ready");
    const said = e2eName("said");

    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Scheduled Job" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-say");
    await page.getByRole("group", { name: "Scheduling Mode" }).getByRole("tab", { name: "No schedule" }).click();
    await page.getByRole("group", { name: "Command *" }).getByRole("textbox").fill(`sh -c "echo ${said} | tr a-z A-Z"`);
    await page.getByRole("button", { name: "Save" }).click();

    const row = page.getByRole("row", { name: /e2e-say/ });
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: "Run Now" }).click();
    await page.getByRole("dialog", { name: "Task created" }).getByRole("button", { name: "View Run" }).click();
    // What it printed, not the command: a command runs without a shell, so
    // the pipe asks for one.
    await expectShownLogs(page, said.toUpperCase());
    // The status and the task's kind are apart in the page: read together. The
    // page asks for the status every 5 seconds once the run's log has ended.
    await expect(page.locator("main")).toContainText(/Done\s*task:sched-job-exec/, RUN_ENDED);
});

// A job on a cron runs on it, unasked. What it prints goes to its run, not to
// the app's log: it writes to the container's own output instead - its first
// process's - where the app's log shows it, once a minute.
test("a scheduled job on a cron runs on it, each minute", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "cron");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");
    const said = e2eName("ticked");

    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Scheduled Job" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-tick");
    await page.getByRole("group", { name: "Scheduling Mode" }).getByRole("tab", { name: "Time-based" }).click();
    await page.getByRole("group", { name: "Cron Expression" }).getByRole("textbox").fill("* * * * *");
    await page
        .getByRole("group", { name: "Command *" })
        .getByRole("textbox")
        .fill(`sh -c "echo ${said}-at-$(date +%H%M) > /proc/1/fd/1"`);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: /e2e-tick/ })).toBeVisible();

    // Two runs, a minute apart, as the log streams them in.
    await expectLogs(page, app, "ready");
    const ran = new RegExp(`${said}-at-(\\d{4})`, "g");
    await expect(async () => {
        const minutes = new Set([...(await copyShownLogs(page)).matchAll(ran)].map(m => m[1]));
        expect(minutes.size, "the minutes it ran in").toBeGreaterThanOrEqual(2);
    }).toPass({ timeout: 200_000, intervals: [10_000] });
});

// A sequence runs its jobs in order: each told how the steps before it went,
// and given the values they wrote to $HIVEPAAS_OUTPUT. A step that fails stops
// it: the steps after it are skipped.
test("a job sequence hands a step's output to the next, and stops at a step that fails", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "sequence");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");
    const version = e2eName("v");
    await createJob(api, app, "e2e-make", `sh -c 'echo "VERSION=${version}" >> "$HIVEPAAS_OUTPUT"; echo made'`);
    await createJob(
        api,
        app,
        "e2e-use",
        `sh -c 'echo "got-$HIVEPAAS_SEQ_OUTPUT_VERSION-after-$HIVEPAAS_SEQ_PREV_STATUS-step-$HIVEPAAS_SEQ_STEP"'`,
    );
    await createJob(api, app, "e2e-fail", "sh -c 'echo failing; exit 3'");

    await newSequence(page, app, "e2e-release", ["e2e-make", "e2e-use"]);
    await runNow(page, "e2e-release");
    // Watched as it runs: each step's lines come as the step runs.
    await expectShownLogs(page, `got-${version}-after-done-step-2`);
    await expect(page.locator("main")).toContainText(/Done\s*task:sched-job-exec/, RUN_ENDED);

    await newSequence(page, app, "e2e-broken", ["e2e-fail", "e2e-use"]);
    await runNow(page, "e2e-broken");
    await expectShownLogs(page, "failing");
    await expect(page.locator("main")).toContainText(/Failed\s*task:sched-job-exec/, { timeout: 60_000 });
    expect(await copyShownLogs(page), "the step after the failed one is skipped").not.toContain("got-");
});

// A job with a trigger runs when the event happens: after a deploy of its
// app, told which event and which deployment.
test("a job triggered after a deploy runs once the deploy is done, told of it", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "trigger");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");

    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Scheduled Job" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-after-deploy");
    await page.getByRole("group", { name: "Scheduling Mode" }).getByRole("tab", { name: "No schedule" }).click();
    await page.getByRole("button", { name: "Add trigger" }).click();
    await expect(page.getByRole("combobox").filter({ hasText: "After a deploy" })).toBeVisible();
    await page
        .getByRole("group", { name: "Command *" })
        .getByRole("textbox")
        .fill(`sh -c 'echo "triggered-$HIVEPAAS_TRIGGER_EVENT-$HIVEPAAS_TRIGGER_DEPLOYMENT"'`);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: /e2e-after-deploy/ })).toBeVisible();

    await redeploy(page, api, app);
    const deployment = (await latestDeployment(api, app))?.id ?? "";
    let run = "";
    await expect
        .poll(
            async () => {
                run = (await jobRuns(api, app)).at(0)?.id ?? "";
                return run;
            },
            { timeout: 120_000, intervals: [3_000] },
        )
        .not.toBe("");
    await page.goto(`${appPage(app, "tasks")}${run}/`);
    await expectShownLogs(page, `triggered-post-deploy-${deployment}`);
});

// A job fails as it is set to: retried as many times as asked before it is
// failed, and stopped at its timeout.
test("a failing job is retried as many times as asked, and a slow one stops at its timeout", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "retry");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");

    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Scheduled Job" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-flaky");
    await page.getByRole("group", { name: "Scheduling Mode" }).getByRole("tab", { name: "No schedule" }).click();
    await page.getByRole("group", { name: "Retry" }).getByRole("button", { name: "Show Configuration" }).click();
    await page.getByRole("spinbutton", { name: "Max" }).fill("2");
    await page.getByRole("textbox", { name: "Delay", exact: true }).fill("2s");
    await page.getByRole("group", { name: "Command *" }).getByRole("textbox").fill("sh -c 'echo attempt; exit 1'");
    await page.getByRole("button", { name: "Save" }).click();
    await runNow(page, "e2e-flaky");
    // A run waiting for its retry reads Failed, and its page does not ask
    // again: it is read anew until the last retry has been made.
    await expect(async () => {
        await page.reload();
        await expect(page.locator("main")).toContainText("Retries: 2", { timeout: 3_000 });
    }).toPass({ timeout: 90_000, intervals: [5_000] });
    await expect(page.locator("main")).toContainText(/Failed\s*task:sched-job-exec/);
    await expect(page.locator("main")).not.toContainText("Retry At");

    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Scheduled Job" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("e2e-slow");
    await page.getByRole("group", { name: "Scheduling Mode" }).getByRole("tab", { name: "No schedule" }).click();
    await page.getByRole("group", { name: "Timeout" }).getByRole("textbox").fill("3s");
    await page.getByRole("group", { name: "Command *" }).getByRole("textbox").fill("sleep 60");
    await page.getByRole("button", { name: "Save" }).click();
    await runNow(page, "e2e-slow");
    // Well before the command's own minute is up.
    await expect(page.locator("main")).toContainText(/Failed\s*task:sched-job-exec/, { timeout: 30_000 });
});

// A job turned off does not run on its schedule; turned on again, it does.
test("a job turned off does not run on its cron, and runs again once turned on", async ({ page, api, cleanup }) => {
    test.setTimeout(420_000);
    const app = await appIn(api, cleanup, "cron-off");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");
    const said = e2eName("ticked");
    await createJob(api, app, "e2e-tick", `sh -c "echo ${said}-at-$(date +%H%M) > /proc/1/fd/1"`, {
        schedule: { cronExpr: "* * * * *" },
    });
    const ran = new RegExp(`${said}-at-(\\d{4})`, "g");
    const minutes = async () => {
        await expectLogs(page, app, "ready");
        return new Set([...(await copyShownLogs(page)).matchAll(ran)].map(m => m[1]));
    };
    await expect(async () => expect((await minutes()).size).toBeGreaterThanOrEqual(1)).toPass({
        timeout: 90_000,
        intervals: [10_000],
    });

    await setJobStatus(page, app, "e2e-tick", "Disabled");
    // What ran before it was turned off, and perhaps a run under way then.
    await page.waitForTimeout(5_000);
    const before = await minutes();
    await page.waitForTimeout(75_000);
    expect(await minutes(), "no run while it is off").toEqual(before);

    await setJobStatus(page, app, "e2e-tick", "Active");
    await expect(async () => expect((await minutes()).size).toBeGreaterThan(before.size)).toPass({
        timeout: 90_000,
        intervals: [10_000],
    });
});

// jobRuns are the runs of the app's scheduled jobs, the newest first.
async function jobRuns(api: APIRequestContext, app: App): Promise<{ id: string; status: string }[]> {
    const res = await api.get(`${appPath(app)}/tasks`);
    const tasks = ((await res.json()) as { data: { id: string; type: string; status: string }[] }).data;
    return tasks.filter(task => task.type === "task:sched-job-exec");
}

// setJobStatus turns a job on or off from its row's Change Status.
async function setJobStatus(page: Page, app: App, name: string, status: "Active" | "Disabled"): Promise<void> {
    await page.goto(appPage(app, "sched-jobs"));
    await page
        .getByRole("row", { name: new RegExp(name) })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: "Change Status" }).click();
    const dialog = page.getByRole("dialog", { name: "Change status" });
    await dialog.getByRole("tab", { name: status }).click();
    await dialog.getByRole("button", { name: "Save" }).click();
    await expect(dialog).toBeHidden();
}

// newSequence makes a job sequence of the app's jobs, in the order given, from
// its Scheduled Jobs.
async function newSequence(page: Page, app: App, name: string, steps: string[]): Promise<void> {
    await page.goto(appPage(app, "sched-jobs"));
    await page.getByRole("button", { name: "New Job Sequence" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill(name);
    for (const step of steps) {
        await page.getByRole("combobox", { name: /^Jobs \*/ }).click();
        await page.getByRole("option", { name: step, exact: true }).click();
        await page.getByRole("button", { name: "Add step" }).click();
    }
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: new RegExp(name) })).toBeVisible();
}

// runNow runs the job from its row, and opens the run.
async function runNow(page: Page, name: string): Promise<void> {
    await page
        .getByRole("row", { name: new RegExp(name) })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: "Run Now" }).click();
    await page.getByRole("dialog", { name: "Task created" }).getByRole("button", { name: "View Run" }).click();
    await page.waitForURL(/\/tasks\/\w+\/?$/);
}
