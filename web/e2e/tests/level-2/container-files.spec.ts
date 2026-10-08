import type { Page } from "@playwright/test";
import { randomBytes } from "node:crypto";

import { appPath, deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 240_000 });

// openUpload opens the terminal's Upload dialog, which sends to the container
// the shell runs in.
async function openUpload(page: Page) {
    await page.getByRole("button", { name: "Upload", exact: true }).click();
    const dialog = page.getByRole("dialog", { name: "Import Files to Container" });
    await expect(dialog).toBeVisible();
    return dialog;
}

test("a file uploaded from the terminal reaches the container whole, over a stream", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "upload");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");

    await page.goto(appPage(app, "terminal"));
    await page.getByRole("button", { name: "Connect" }).click();
    // The shell is opened in the container before the socket is.
    await expect(page.getByText("connected", { exact: true })).toBeVisible({ timeout: 30_000 });

    // Larger than a piece the dashboard sends at a time, so it goes in several.
    const name = `${e2eName("upload")}.bin`;
    const content = randomBytes(3 * 1024 * 1024 + 123);

    let dialog = await openUpload(page);
    await dialog
        .locator('input[type="file"]')
        .setInputFiles({ name, mimeType: "application/octet-stream", buffer: content });
    // A path that ends in a slash is a directory, the file keeping its name in it.
    await dialog.getByRole("group", { name: "Destination Path" }).getByRole("textbox").fill("/tmp/");
    const stream = page.waitForEvent("websocket", {
        predicate: ws => ws.url().includes("/container/file-upload/stream"),
        timeout: 30_000,
    });
    let binaryFrames = 0;
    void stream.then(ws =>
        ws.on("framesent", frame => {
            if (typeof frame.payload !== "string") binaryFrames++;
        }),
    );
    await dialog.getByRole("button", { name: "Upload", exact: true }).click();
    await stream;
    await expect(page.locator('[data-sonner-toast][data-type="success"]')).toBeVisible({ timeout: 60_000 });
    await expect(dialog).toBeHidden();
    expect(binaryFrames, "pieces sent").toBeGreaterThan(1);

    const res = await api.get(`${appPath(app)}/container/file-download`, {
        params: { path: `/tmp/${name}`, isDir: false },
    });
    expect(res.status(), await res.text().catch(() => "")).toBe(200);
    expect(Buffer.compare(await res.body(), content), "the file as it was sent").toBe(0);

    // Where nothing is, the container refuses: the dialog stays, and says so.
    dialog = await openUpload(page);
    await dialog
        .locator('input[type="file"]')
        .setInputFiles({ name, mimeType: "application/octet-stream", buffer: content });
    await dialog.getByRole("group", { name: "Destination Path" }).getByRole("textbox").fill("/e2e/no/such/dir/");
    await dialog.getByRole("button", { name: "Upload", exact: true }).click();
    await expect(page.getByText("Failed to upload container file").first()).toBeVisible({ timeout: 60_000 });
    await expect(dialog.getByRole("button", { name: "Upload", exact: true })).toBeEnabled();
});
