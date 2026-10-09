import type { Page } from "@playwright/test";
import { randomBytes } from "node:crypto";

import { appPath, deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { env } from "../../support/env";
import { e2eName, expect, test } from "../../support/fixtures";
import { signIn } from "../../support/sign-in";

test.describe.configure({ timeout: 240_000 });
// Signs in afresh, not with the shared session: the test refreshes its session,
// which the others' would then be refreshed from no more.
test.use({ storageState: { cookies: [], origins: [] } });

// EXPIRED_TOKEN is an access token that expired long ago: the dashboard reads
// its expiry, the server never sees it once refreshed.
const EXPIRED_TOKEN = [{ alg: "HS256", typ: "JWT" }, { exp: 1 }]
    .map(part => Buffer.from(JSON.stringify(part)).toString("base64url"))
    .concat("expired")
    .join(".");

// openUpload opens the terminal's Upload dialog, which sends to the container
// the shell runs in.
async function openUpload(page: Page) {
    await page.getByRole("button", { name: "Upload", exact: true }).click();
    const dialog = page.getByRole("dialog", { name: "Import Files to Container" });
    await expect(dialog).toBeVisible();
    return dialog;
}

test("a file uploaded from the terminal reaches the container whole, over a stream", async ({ page, api, cleanup }) => {
    await page.goto("/");
    await signIn(page, env.username, env.password);
    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
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

    // Without its slash, a path names the file: one that is a directory is not
    // replaced by it, and the dialog says what to do instead.
    await dialog.getByRole("group", { name: "Destination Path" }).getByRole("textbox").fill("/tmp");
    await dialog.getByRole("button", { name: "Upload", exact: true }).click();
    await expect(uploadFailure(page)).toContainText("end the path with /", { timeout: 60_000 });
    await expect(page.locator("[data-sonner-toast]")).toHaveCount(0, { timeout: 30_000 });

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
    // What the server answered, not that the connection closed.
    await expect(uploadFailure(page)).toContainText(/\/e2e\/no\/such\/dir/, { timeout: 60_000 });
    await expect(dialog.getByRole("button", { name: "Upload", exact: true })).toBeEnabled();
    // Gone on its own, for the next to be told apart.
    await expect(page.locator("[data-sonner-toast]")).toHaveCount(0, { timeout: 30_000 });

    // A session expired while the dialog was open is refreshed, as for any
    // request: a websocket given the expired token would be refused, and a
    // browser is told nothing about a refused upgrade.
    await page.evaluate(token => localStorage.setItem("token", token), EXPIRED_TOKEN);
    await dialog.getByRole("group", { name: "Destination Path" }).getByRole("textbox").fill("/tmp/");
    await dialog.getByRole("button", { name: "Upload", exact: true }).click();
    await expect(page.locator('[data-sonner-toast][data-type="success"]')).toBeVisible({ timeout: 60_000 });
    await expect(dialog).toBeHidden();
});

// uploadFailure is where the dialog's failure is told: a toast, or an alert for
// a grave one.
function uploadFailure(page: Page) {
    return page
        .locator('[data-sonner-toast], [role="alertdialog"]')
        .filter({ hasText: "Failed to upload container file" })
        .first();
}

test("an upload is cancelled from its dialog", async ({ page, api, cleanup }) => {
    await page.goto("/");
    await signIn(page, env.username, env.password);
    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    const app = await appIn(api, cleanup, "upload-cancel");
    await deployImage(api, app, BUSYBOX, "sh -c 'echo ready; exec sleep 3600'");
    await deployed(api, app);
    await expectLogs(page, app, "ready");
    await page.goto(appPage(app, "terminal"));
    await page.getByRole("button", { name: "Connect" }).click();
    await expect(page.getByText("connected", { exact: true })).toBeVisible({ timeout: 30_000 });

    // Slow enough to be cancelled on its way.
    const cdp = await page.context().newCDPSession(page);
    await cdp.send("Network.enable");
    await cdp.send("Network.emulateNetworkConditions", {
        offline: false,
        latency: 0,
        downloadThroughput: -1,
        uploadThroughput: 512 * 1024,
    });
    const name = `${e2eName("cancelled")}.bin`;
    const dialog = await openUpload(page);
    await dialog
        .locator('input[type="file"]')
        .setInputFiles({ name, mimeType: "application/octet-stream", buffer: randomBytes(16 * 1024 * 1024) });
    await dialog.getByRole("group", { name: "Destination Path" }).getByRole("textbox").fill("/tmp/");
    await dialog.getByRole("button", { name: "Upload", exact: true }).click();
    await expect(dialog.getByRole("progressbar", { name: "Upload progress" })).toBeVisible({ timeout: 30_000 });

    await dialog.getByRole("button", { name: "Cancel upload" }).click();

    await expect(page.getByText("Upload cancelled")).toBeVisible();
    await expect(dialog.getByRole("button", { name: "Upload", exact: true })).toBeEnabled();
});
