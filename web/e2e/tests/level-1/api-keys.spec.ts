import { type APIRequestContext, request } from "@playwright/test";

import { createProject, deleteApiKeysNamed, deleteProject, findProject } from "../../support/api";
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
        // The row leaves the table before the server is done: wait for the
        // server, not the row.
        const deleted = page.waitForResponse(
            res => res.request().method() === "DELETE" && res.url().includes("/settings/api-keys/"),
        );
        await page.getByRole("dialog", { name: "Delete Item" }).getByRole("button", { name: "Delete" }).click();
        expect((await deleted).ok()).toBe(true);

        await expect(page.getByRole("row").filter({ hasText: name })).toHaveCount(0);
        const refused = await withKey.get("projects");
        expect(refused.status()).toBe(401);
        expect(((await refused.json()) as { code: string }).code).toBe("ERR_API_KEY_INVALID");
    } finally {
        await withKey.dispose();
    }
});

test("a key is asked for when it expires before anything is sent", async ({ page }) => {
    const sent: string[] = [];
    page.on("request", req => {
        if (req.method() === "POST" && req.url().includes("/settings/api-keys")) sent.push(req.url());
    });
    await page.goto("/current-user/api-keys/create/");

    await page.getByRole("textbox", { name: "Enter API key name" }).fill(e2eName("api-key-undated"));
    await page.getByRole("button", { name: "Create Key" }).click();

    await expect(page.getByRole("group", { name: /^Access Expiration/ })).toContainText("Choose when the key expires");
    expect(sent).toEqual([]);
});

interface Actions {
    read: boolean;
    write: boolean;
    execute: boolean;
    delete: boolean;
}

// keyWith makes an API key of the tests' user allowed only the actions given,
// and answers a client that calls the API with it.
async function keyWith(api: APIRequestContext, name: string, accessAction: Actions): Promise<APIRequestContext> {
    const expireAt = new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString();
    const res = await api.post("users/current/settings/api-keys", {
        data: { name, accessAction, capabilities: [], default: false, inheritable: false, expireAt },
    });
    expect(res.ok(), `creating API key ${name}: ${res.status()} ${await res.text()}`).toBe(true);
    const key = ((await res.json()) as { data: { keyId: string; secretKey: string } }).data;
    return request.newContext({
        baseURL: `${env.baseURL}/api/`,
        extraHTTPHeaders: { "HIVEPAAS-API-KEY-ID": key.keyId, "HIVEPAAS-API-SECRET-KEY": key.secretKey },
    });
}

// A key does what its actions allow and nothing more, though its user may do
// it all: a read-only key reads and is refused a write, a key without delete is
// refused a deletion - each refused as the API refuses an action not allowed,
// with nothing made or removed.
test("an API key is refused what its actions do not allow", async ({ api, cleanup }) => {
    const readOnly = e2eName("api-key-read");
    const noDelete = e2eName("api-key-no-delete");
    cleanup(() => deleteApiKeysNamed(api, readOnly));
    cleanup(() => deleteApiKeysNamed(api, noDelete));
    const project = await createProject(api, e2eName("key-actions"));
    cleanup(() => deleteProject(api, project.id));

    const reader = await keyWith(api, readOnly, { read: true, write: false, execute: false, delete: false });
    const writer = await keyWith(api, noDelete, { read: true, write: true, execute: true, delete: false });
    try {
        expect((await reader.get("projects")).status(), "a read-only key reads").toBe(200);

        const made = e2eName("made-by-read-key");
        cleanup(async () => {
            const stray = await findProject(api, made);
            if (stray) await deleteProject(api, stray.id);
        });
        const write = await reader.post("projects", {
            data: { name: made, status: "active", envs: [{ name: "development", color: "#a855f7" }] },
        });
        expect(write.status(), "a read-only key is refused a write").toBe(401);
        expect(((await write.json()) as { code: string }).code).toBe("ERR_UNAUTHORIZED");
        expect(await findProject(api, made), "nothing was made").toBeUndefined();

        const removal = await writer.delete(`projects/${project.id}`);
        expect(removal.status(), "a key without delete is refused a deletion").toBe(401);
        expect(((await removal.json()) as { code: string }).code).toBe("ERR_UNAUTHORIZED");
        expect(await findProject(api, project.name), "nothing was removed").toBeDefined();
    } finally {
        await reader.dispose();
        await writer.dispose();
    }
});
