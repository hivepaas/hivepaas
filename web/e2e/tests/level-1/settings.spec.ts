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

    test("an interval that is not one is refused, and not kept", async ({ page, api, cleanup }) => {
        const saved = await cleanupSettings(api);
        cleanup(() => restoreCleanupSettings(api, saved));
        await page.goto("/settings/data-cleanup/configuration/");
        const interval = page.getByRole("group", { name: "Scheduling Interval" }).getByRole("textbox");
        const before = await interval.inputValue();

        await interval.fill("every day");
        await page.getByRole("button", { name: "Save" }).click();

        await expect(page.getByText("Failed to update system cleanup settings")).toBeVisible();
        await page.reload();
        await expect(interval).toHaveValue(before);
    });
});
