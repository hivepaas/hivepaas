import { deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs, expectShownLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 240_000 });

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
