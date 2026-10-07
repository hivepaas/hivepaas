import { expect, test } from "../../support/fixtures";

// The projects list leaves out the project HivePaaS runs in: System › HivePaaS
// › Actions is the way to its apps, for their logs, metrics and instances.
test("View System Apps opens the apps HivePaaS runs itself", async ({ page }) => {
    await page.goto("/system/hivepaas/actions/");
    await expect(
        page.getByText("Changing their settings directly can break HivePaaS.", { exact: false }),
    ).toBeVisible();

    await page.getByRole("link", { name: "View System Apps" }).click();

    await expect(page).toHaveURL(/\/projects\/[^/]+\/apps\/$/);
    // Two that every installation runs, whatever else it does.
    await expect(page.getByRole("row").filter({ hasText: "traefik" })).toBeVisible();
    await expect(page.getByRole("row").filter({ hasText: "worker" })).toBeVisible();
});
