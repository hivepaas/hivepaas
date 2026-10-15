import type { APIRequestContext, Page } from "@playwright/test";

import { createVolume, deleteVolume, deployImage } from "../../support/api";
import { appIn, appPage, deployed } from "../../support/apps";
import { type Cleanup, e2eName, expect, test } from "../../support/fixtures";

// The first scan on a volume fetches Trivy's image and its DB, then reads the
// image: room for all three. Scanning is one switch for the installation: the
// tests take turns with it.
test.describe.configure({ timeout: 480_000, mode: "serial" });

// alpine:3.10 is past its support, and its apk-tools has CVE-2021-36159,
// CRITICAL, fixed in 2.10.7-r0: a finding the DB will not drop.
const OLD_ALPINE = "alpine:3.10";
const CVE = "CVE-2021-36159";

const SETTINGS = "/system/image-scanning/configuration/";
const POLL = { timeout: 300_000, intervals: [5_000] };

interface ImageScanSettings {
    updateVer: number;
    enabled: boolean;
    volume?: { id: string };
    dbRepository: string;
    javaDbRepository: string;
    schedule: { interval: string; initialTime: string };
    notification: { successUseDefault: boolean; failureUseDefault: boolean };
    cpuLimit: number;
    memoryLimit: string;
}

async function imageScanSettings(api: APIRequestContext): Promise<ImageScanSettings> {
    const res = await api.get("system/settings/image-scan");
    expect(res.ok(), await res.text()).toBe(true);
    return ((await res.json()) as { data: ImageScanSettings }).data;
}

// putSettings saves the settings as they are, switched on with a volume or off.
async function putSettings(api: APIRequestContext, enabled: boolean, volumeId?: string): Promise<void> {
    const current = await imageScanSettings(api);
    const res = await api.put("system/settings/image-scan", {
        data: {
            updateVer: current.updateVer,
            enabled,
            volume: volumeId ? { id: volumeId } : (current.volume ?? null),
            dbRepository: current.dbRepository,
            javaDbRepository: current.javaDbRepository,
            schedule: { interval: current.schedule.interval },
            notification: current.notification,
            cpuLimit: current.cpuLimit,
            memoryLimit: current.memoryLimit,
        },
    });
    expect(res.ok(), await res.text()).toBe(true);
}

// switchOff puts scanning back off, as the installation had it.
async function switchOff(api: APIRequestContext): Promise<void> {
    if ((await imageScanSettings(api)).enabled) {
        await putSettings(api, false);
    }
}

// scanVolume makes the volume a test's scans keep Trivy's data on, pinned to
// the node HivePaaS runs on; it goes once scanning is off.
async function scanVolume(api: APIRequestContext, cleanup: Cleanup): Promise<{ id: string; name: string }> {
    const name = e2eName(`trivy-${Date.now().toString(36)}`);
    const id = await createVolume(api, name);
    cleanup(async () => {
        await switchOff(api);
        await deleteVolume(api, id);
    });
    return { id, name };
}

// criticalShown waits for the Security tab to count the CRITICAL finding, and
// scans again when the first scan came too early - before the server had its
// DB.
async function criticalShown(page: Page): Promise<void> {
    await expect
        .poll(async () => {
            await page.reload();
            const scan = page.getByRole("button", { name: "Scan now" });
            const failed =
                (await page.getByText("Scan failed").count()) > 0 || (await page.getByText("Not scanned:").count()) > 0;
            if (failed && (await scan.isEnabled())) {
                await scan.click();
            }
            return page.getByTestId("severity-critical").textContent();
        }, POLL)
        .toMatch(/Critical\s*1\s*1 fixable/);
}

test("image scanning: switched on with a volume, an app's image is scanned after its deployment, on the volume's node; its fixable CRITICAL and its unsupported OS show on its Security tab, its header and Needs attention", async ({
    page,
    api,
    cleanup,
}) => {
    cleanup(() => switchOff(api));
    const volume = await scanVolume(api, cleanup);

    await test.step("switched on, a volume is required; the one chosen says where scans run", async () => {
        await page.goto(SETTINGS);
        const enabled = page.getByRole("checkbox", { name: "Enabled" });
        if (!(await enabled.isChecked())) {
            await enabled.click();
        }
        await expect(page.getByText("nothing runs between scans")).toBeVisible();
        const volumeField = page.getByRole("group", { name: /^Data volume/ });
        await volumeField.getByRole("combobox").click();
        await page.getByRole("option", { name: volume.name }).click();
        await expect(page.getByTestId("image-scan-where")).toContainText("the node this volume is on");
        await page.getByRole("button", { name: "Save" }).click();
        await expect(page.getByText("Image scanning settings updated")).toBeVisible();

        const status = page.getByTestId("image-scanner-status");
        await expect(status.getByRole("group", { name: "Data volume" })).toContainText(volume.name);
        await expect(status.getByRole("group", { name: "Scans run" })).toContainText("where the volume is");
        await expect(status.getByText(/aquasec\/trivy:/)).toBeVisible();
    });

    const app = await appIn(api, cleanup, "image-scan");
    await deployImage(api, app, OLD_ALPINE, "sleep 3600");
    await deployed(api, app);

    await test.step("the Security tab shows the fixable CRITICAL and the OS past its support", async () => {
        await page.goto(appPage(app, "security"));
        await criticalShown(page);

        const row = page.getByTestId("finding-row").filter({ hasText: CVE });
        await expect(row).toBeVisible();
        await expect(row).toContainText("apk-tools");
        await expect(row).toContainText("2.10.6-r0");
        await expect(row).toContainText("2.10.7-r0");
        await expect(row.getByRole("link", { name: CVE })).toHaveAttribute("href", /cve-2021-36159/);

        const image = page.getByTestId("scanned-image");
        await expect(image).toContainText(OLD_ALPINE);
        await expect(image).toContainText("is past its support");
        await expect(image).toContainText("vulnerability DB of");
    });

    await test.step("the header's badge counts it, and opens the Security tab", async () => {
        await page.goto(appPage(app, "general"));
        const badge = page.getByTestId("app-image-scan-badge");
        await expect(badge).toHaveText(/1 critical/, POLL);
        await badge.click();
        await expect(page).toHaveURL(/\/security\/$/);
    });

    await test.step("Needs attention lists the app, and leads to its findings", async () => {
        // The list loads after the page: each try waits for it a little.
        await expect(async () => {
            await page.goto("/home/");
            await expect(page.getByText("runs a fixable critical vulnerability").first()).toBeVisible({
                timeout: 10_000,
            });
        }).toPass(POLL);
        const item = page
            .getByRole("listitem")
            .filter({ hasText: `${app.name} runs a fixable critical vulnerability` });
        await expect(item.getByRole("link", { name: "View findings" }).first()).toBeVisible();
    });

    await test.step("System > Image Scanning > Apps lists the app, the worst first", async () => {
        await page.goto("/system/image-scanning/apps/");
        const row = page
            .getByTestId("image-scan-app-row")
            .filter({ hasText: app.name })
            .filter({ hasText: OLD_ALPINE });
        await expect(row.first()).toBeVisible();
        await expect(row.first()).toContainText("unsupported");
    });
});

test("image scanning: a finding ignored with a reason leaves the counts, the badge and the list for Ignored; removed, it counts again; switched off, the results stay", async ({
    page,
    api,
    cleanup,
}) => {
    cleanup(() => switchOff(api));
    const volume = await scanVolume(api, cleanup);
    await putSettings(api, true, volume.id);

    const app = await appIn(api, cleanup, "image-ignore");
    await deployImage(api, app, OLD_ALPINE, "sleep 3600");
    await deployed(api, app);
    await page.goto(appPage(app, "security"));
    await criticalShown(page);

    await test.step("ignored for the app, with a reason, it leaves the counts and the badge", async () => {
        await page.getByTestId("finding-row").filter({ hasText: CVE }).getByRole("button", { name: "Ignore…" }).click();
        const dialog = page.getByRole("dialog");
        await expect(dialog.getByRole("heading", { name: `Ignore ${CVE}` })).toBeVisible();
        // A reason is required.
        await expect(dialog.getByRole("button", { name: "Ignore", exact: true })).toBeDisabled();
        await dialog.getByRole("textbox", { name: "Reason" }).fill("apk never runs in the container");
        await dialog.getByRole("button", { name: "Ignore", exact: true }).click();
        await expect(page.getByText("Vulnerability ignored")).toBeVisible();

        await expect(page.getByTestId("severity-critical")).toContainText("0");
        await expect(page.getByTestId("finding-row").filter({ hasText: CVE })).toHaveCount(0);
        await page.getByRole("button", { name: "Ignored (1)" }).click();
        const ignored = page.getByTestId("ignored-row").filter({ hasText: CVE });
        await expect(ignored).toContainText("This app");
        await expect(ignored).toContainText("apk never runs in the container");
        await expect(ignored).toContainText("admin");

        await page.goto(appPage(app, "general"));
        await expect(page.getByTestId("app-image-scan-badge")).toHaveCount(0);
    });

    await test.step("removed, it counts again", async () => {
        await page.goto(appPage(app, "security"));
        await page.getByRole("button", { name: "Ignored (1)" }).click();
        await page.getByTestId("ignored-row").filter({ hasText: CVE }).getByRole("button", { name: "Remove" }).click();
        await page.getByRole("button", { name: "Remove", exact: true }).last().click();
        await expect(page.getByText("Ignore removed")).toBeVisible();
        await expect(page.getByTestId("severity-critical")).toContainText(/Critical\s*1/);
    });

    await test.step("switched off, nothing is scanned; the app's results stay, said to be the last scan's", async () => {
        await page.goto(SETTINGS);
        await page.getByRole("checkbox", { name: "Enabled" }).click();
        await expect(page.getByText("Switched off, nothing is scanned")).toBeVisible();
        await page.getByRole("button", { name: "Save" }).click();
        await expect(page.getByText("Image scanning settings updated")).toBeVisible();
        await expect.poll(async () => (await imageScanSettings(api)).enabled).toBe(false);

        await page.goto(appPage(app, "security"));
        await expect(page.getByText("Image scanning is switched off")).toBeVisible();
        await expect(page.getByTestId("finding-row").filter({ hasText: CVE })).toBeVisible();
        await expect(page.getByRole("button", { name: "Scan now" })).toHaveCount(0);
    });
});
