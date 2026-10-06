import { cleanupSettings, restoreCleanupSettings } from "../../support/api";
import { expect, test } from "../../support/fixtures";

// The settings are the installation's own, one of each: these tests take
// turns, and put back what they found.
test.describe.configure({ mode: "serial" });

test.describe("Data Cleanup settings", () => {
    test("what is saved is kept", async ({ page, api, cleanup }) => {
        const saved = await cleanupSettings(api);
        cleanup(() => restoreCleanupSettings(api, saved));
        await page.goto("/settings/data-cleanup/configuration/");
        const retention = page.getByRole("group", { name: "Audit Logs Retention" }).getByRole("textbox");
        const syncApps = page.getByRole("group", { name: "Sync With Settings" }).getByRole("checkbox");

        await retention.fill("120d");
        await syncApps.uncheck();
        await page.getByRole("button", { name: "Save" }).click();

        await expect(page.getByText(/settings updated/i)).toBeVisible();
        await page.reload();
        await expect(retention).toHaveValue("120d");
        await expect(syncApps).not.toBeChecked();
    });

    test("an interval or a retention that is not a duration is caught in the form", async ({ page }) => {
        const sent: string[] = [];
        page.on("request", req => {
            if (req.method() === "PUT" && req.url().includes("/system/settings/cleanup")) sent.push(req.url());
        });
        await page.goto("/settings/data-cleanup/configuration/");
        const interval = page.getByRole("group", { name: "Scheduling Interval" });
        const retention = page.getByRole("group", { name: "Audit Logs Retention" });

        await interval.getByRole("textbox").fill("every day");
        await retention.getByRole("textbox").fill("three months");
        await page.getByRole("button", { name: "Save" }).click();

        await expect(interval).toContainText("Use a duration such as 1d, 12h or 1h30m");
        await expect(retention).toContainText("Use a duration such as 1d, 12h or 1h30m");
        expect(sent).toEqual([]);
    });
});
