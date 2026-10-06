import { request } from "@playwright/test";

import { deleteApiKeysNamed } from "../../support/api";
import { env } from "../../support/env";
import { e2eName, expect, test } from "../../support/fixtures";

test("an API key answers once made, shows its secret once, and stops once deleted", async ({ page, api, cleanup }) => {
    const name = e2eName("api-key");
    cleanup(() => deleteApiKeysNamed(api, name));
    await page.goto("/current-user/api-keys/create/");

    await page.getByRole("textbox", { name: "Enter API key name" }).fill(name);
    await page.getByRole("button", { name: "Select expiration date" }).click();
    const calendar = page.getByRole("dialog");
    await calendar.getByRole("button", { name: "Go to the Next Month" }).click();
    await calendar.getByRole("grid").getByRole("button", { disabled: false }).first().click();
    await page.keyboard.press("Escape");
    await page.getByRole("button", { name: "Create Key" }).click();

    await expect(page.getByText("API key created successfully")).toBeVisible();
    await expect(page.getByText("The secret key will not be stored on the server")).toBeVisible();
    const keyId = await page.getByText("Key ID:").locator("xpath=following-sibling::p[1]").innerText();
    const secret = await page.getByText("Secret key:").locator("xpath=following-sibling::p[1]").innerText();

    const withKey = await request.newContext({
        baseURL: `${env.baseURL}/api/`,
        extraHTTPHeaders: { "HIVEPAAS-API-KEY-ID": keyId.trim(), "HIVEPAAS-API-SECRET-KEY": secret.trim() },
    });
    try {
        expect((await withKey.get("projects")).status()).toBe(200);

        await page.goto("/current-user/api-keys/");
        await page.getByRole("row").filter({ hasText: name }).getByRole("button", { name: "Actions menu" }).click();
        await page.getByRole("menu").getByRole("button", { name: "Delete" }).click();
        await page.getByRole("dialog").getByRole("button", { name: "Delete" }).click();

        await expect(page.getByRole("row").filter({ hasText: name })).toHaveCount(0);
        // Refused as a key it does not know. (It answers 412, not 401.)
        const refused = await withKey.get("projects");
        expect(refused.ok()).toBe(false);
        expect(((await refused.json()) as { code: string }).code).toBe("ERR_API_KEY_INVALID");
    } finally {
        await withKey.dispose();
    }
});
