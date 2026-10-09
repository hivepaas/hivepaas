import { deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, copyShownLogs, deployed, expectLogs, expectShownLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

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
    // The status and the task's kind are apart in the page: read together.
    await expect(page.locator("main")).toContainText(/Done\s*task:sched-job-exec/);
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
