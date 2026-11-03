import type { APIRequestContext, Page } from "@playwright/test";

import { expect, test } from "../../support/fixtures";
import { canSignLicenses, issueKey, noSigner, readLicense, usageOf } from "../../support/license";

// System › License, on the Colony development key env/up.sh installs. Nothing
// here changes the key in force - every other test counts on it, and these run
// beside them: a key refused is not stored. The tests that change it are
// tests/level-2/license-on/, run alone.

const LICENSE_PAGE = "/system/license/overview/";

// pasteKey pastes a key into Install a key and installs it: replacing the
// Colony key in force is asked first.
async function pasteKey(page: Page, key: string): Promise<void> {
    const form = page.getByRole("region", { name: "Install a key" });
    await form.getByRole("textbox", { name: "License key" }).fill(key);
    await form.getByRole("button", { name: "Install key" }).click();
    const confirm = page.getByRole("alertdialog", { name: "Replace the license key?" });
    await expect(confirm).toContainText("This key replaces the Colony key in force, whatever its plan.");
    await confirm.getByRole("button", { name: "Replace key" }).click();
}

// keyFieldError is what the form says beside the key box.
function keyFieldError(page: Page) {
    return page.getByRole("region", { name: "Install a key" }).locator('[data-slot="field-error"]');
}

// The key in force is the one env/up.sh installed, unchanged: shown, and as
// the API reads it.
async function expectColonyKept(page: Page, api: APIRequestContext, issuedAt: string | undefined): Promise<void> {
    const after = await readLicense(api);
    expect(`${after.plan} ${after.status} ${after.licenseId}`).toBe("colony valid lic_e2e");
    expect(after.issuedAt, "the same key").toBe(issuedAt);
    const plan = page.getByRole("region", { name: "Plan", exact: true });
    await expect(plan.getByText("Colony", { exact: true })).toBeVisible();
    await expect(plan.getByText("Active", { exact: true })).toBeVisible();
    // Said beside the box, not in a toast.
    await expect(page.locator('[data-sonner-toast][data-type="error"]')).toHaveCount(0);
}

test.describe("System › License", () => {
    test.use({ permissions: ["clipboard-read", "clipboard-write"] });

    // 1.19
    test("shows the Colony development key in force, its usage, the portal's pages, and the installation ID to copy", async ({
        page,
        api,
    }) => {
        const license = await readLicense(api);
        expect(`${license.plan} ${license.status}`, "the key env/up.sh installed").toBe("colony valid");

        await page.goto(LICENSE_PAGE);
        const plan = page.getByRole("region", { name: "Plan", exact: true });
        await expect(plan.getByText("Colony", { exact: true })).toBeVisible();
        await expect(plan.getByText("Active", { exact: true })).toBeVisible();
        await expect(plan.getByText("Development key", { exact: true })).toBeVisible();
        await expect(plan.getByRole("group", { name: "Licensed to" })).toContainText("HivePaaS e2e");
        await expect(plan.getByRole("group", { name: "License ID" })).toContainText("lic_e2e");
        await expect(plan.getByRole("group", { name: "Paid until" })).toBeVisible();

        // The portal's pages, for this installation, in a new tab; a trial only
        // while no key was ever installed.
        const upgrade = plan.getByRole("link", { name: "Upgrade" });
        await expect(upgrade).toHaveAttribute("href", license.portal?.pricingUrl ?? "no pricing page");
        await expect(upgrade).toHaveAttribute("href", new RegExp(`installationId=${license.installationId}`));
        await expect(upgrade).toHaveAttribute("target", "_blank");
        await expect(plan.getByRole("link", { name: "Your licenses" })).toHaveAttribute(
            "href",
            license.portal?.accountUrl ?? "no account page",
        );
        await expect(plan.getByRole("link", { name: "Start a 14-day trial" })).toHaveCount(0);

        // Each limit against Colony's, the noun going with the max; the count of
        // users and apps moves as the tests beside this one make and remove them.
        const usage = page.getByRole("region", { name: "Usage" });
        const nodes = usageOf(license, "nodes");
        await expect(usage.getByRole("meter", { name: "Nodes" })).toHaveAttribute(
            "aria-valuetext",
            `${nodes.used} of ${nodes.max} ${nodes.max === 1 ? "node" : "nodes"}`,
        );
        await expect(usage.getByRole("meter", { name: "Users" })).toHaveAttribute(
            "aria-valuetext",
            new RegExp(`^\\d+ of ${usageOf(license, "users").max} users$`),
        );
        await expect(usage.getByRole("group", { name: "Apps", exact: true })).toContainText(/\d+ apps? · no limit/);
        await expect(usage.getByRole("group", { name: "Google, Microsoft and OpenID Connect sign-in" })).toContainText(
            "Included",
        );

        // The ID a key is bought for, copied as it is.
        const installation = page.getByRole("region", { name: "Installation", exact: true });
        await expect(installation.getByRole("group", { name: "Installation ID" })).toContainText(
            license.installationId,
        );
        await installation.getByRole("button", { name: "Copy the installation id" }).click();
        await expect(page.locator('[data-sonner-toast][data-type="success"]')).toContainText("Installation ID copied");
        expect(await page.evaluate(() => navigator.clipboard.readText())).toBe(license.installationId);
    });

    // 1.20 - the throwaway's portal, http://localhost:4100 inside dind, never
    // answers.
    test("Check for a newer key says the portal could not be reached, and so does the last check", async ({ page }) => {
        await page.goto(LICENSE_PAGE);
        const renewal = page.getByRole("region", { name: "Online renewal" });
        await renewal.getByRole("button", { name: "Check for a newer key" }).click();

        const toast = page.locator('[data-sonner-toast][data-type="error"]');
        await expect(toast).toContainText("Could not check for a newer key");
        await expect(toast).toContainText("The license portal could not be reached");
        const last = renewal.getByRole("group", { name: "Last check" });
        await expect(last).toContainText("The check failed");
        await expect(last).toContainText("The license portal could not be reached");
    });

    // 1.21
    test("a text that is not a key is refused beside the box, and the key in force stays", async ({ page, api }) => {
        const before = await readLicense(api);
        await page.goto(LICENSE_PAGE);

        await pasteKey(page, "not-a-license-key");

        await expect(keyFieldError(page)).toContainText("This license key could not be verified");
        await expectColonyKept(page, api, before.issuedAt);
    });

    // 1.21
    test("another installation's key is refused with this installation's ID, and the key in force stays", async ({
        page,
        api,
    }) => {
        test.skip(!canSignLicenses, noSigner);
        const before = await readLicense(api);
        const otherKey = issueKey({ plan: "hive", installation: "e2e-other-cluster", licenseId: "lic_e2e_other" });
        await page.goto(LICENSE_PAGE);

        await pasteKey(page, otherKey);

        await expect(keyFieldError(page)).toContainText(
            `This license key is for another installation. This installation's ID is ${before.installationId}.`,
        );
        await expectColonyKept(page, api, before.issuedAt);
    });
});
