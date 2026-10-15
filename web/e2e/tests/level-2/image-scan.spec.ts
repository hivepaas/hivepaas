import type { APIRequestContext, Page } from "@playwright/test";
import { readFile } from "node:fs/promises";

import {
    createProject,
    createSettingAt,
    createVolume,
    created,
    deleteProject,
    deleteSettingsNamed,
    deleteVolume,
    deployImage,
    findProject,
} from "../../support/api";
import { appIn, appPage, deployed } from "../../support/apps";
import { type Cleanup, e2eName, expect, test } from "../../support/fixtures";
import { echoHook, slackMessages } from "../../support/notifications";

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
    notification: {
        success?: { id: string } | null;
        successUseDefault: boolean;
        failure?: { id: string } | null;
        failureUseDefault: boolean;
    };
    cpuLimit: number;
    memoryLimit: string;
}

async function imageScanSettings(api: APIRequestContext): Promise<ImageScanSettings> {
    const res = await api.get("system/settings/image-scan");
    expect(res.ok(), await res.text()).toBe(true);
    return ((await res.json()) as { data: ImageScanSettings }).data;
}

// putSettings saves the settings as they are, switched on with a volume or off,
// and with the daily scan's notification when given.
async function putSettings(
    api: APIRequestContext,
    enabled: boolean,
    volumeId?: string,
    notification?: ImageScanSettings["notification"],
): Promise<void> {
    const current = await imageScanSettings(api);
    const res = await api.put("system/settings/image-scan", {
        data: {
            updateVer: current.updateVer,
            enabled,
            volume: volumeId ? { id: volumeId } : (current.volume ?? null),
            dbRepository: current.dbRepository,
            javaDbRepository: current.javaDbRepository,
            schedule: { interval: current.schedule.interval },
            notification: notification ?? current.notification,
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
                (await page.getByText("Scan failed").count()) > 0 ||
                (await page.getByText(/^(Not scanned|Last scan incomplete):/).count()) > 0;
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

interface ScannedComponent {
    component: { key: string; name: string };
    scannedAt: string;
}

// componentScanned waits for the daily scan, run now, to have scanned
// HivePaaS's part of that key.
async function componentScanned(api: APIRequestContext, key: string): Promise<void> {
    const run = await api.post("system/settings/image-scan/exec", { data: {} });
    expect(run.ok(), await run.text()).toBe(true);
    await expect
        .poll(async () => {
            const res = await api.get("system/settings/image-scan/hivepaas");
            const body = (await res.json()) as { data: { components: ScannedComponent[] } };
            return body.data.components.some(c => c.component.key === key);
        }, POLL)
        .toBe(true);
}

test("image scanning, phase 2: a project's Security tab lists its apps; an image's SBOM downloads; HivePaaS's own images are scanned; an ignore goes with its project's spec", async ({
    page,
    api,
    cleanup,
}) => {
    cleanup(() => switchOff(api));
    const volume = await scanVolume(api, cleanup);
    await putSettings(api, true, volume.id);

    const app = await appIn(api, cleanup, "image-p2");
    await deployImage(api, app, OLD_ALPINE, "sleep 3600");
    await deployed(api, app);
    await page.goto(appPage(app, "security"));
    await criticalShown(page);

    await test.step("an image's SBOM downloads as CycloneDX, and names apk-tools", async () => {
        const downloading = page.waitForEvent("download");
        await page.getByTestId("scanned-image").getByRole("button", { name: "Download SBOM" }).click();
        const download = await downloading;
        expect(download.suggestedFilename()).toMatch(/^web-[0-9a-f]{12}\.cdx\.json$/);
        const file = test.info().outputPath("sbom.cdx.json");
        await download.saveAs(file);
        const sbom = JSON.parse(await readFile(file, "utf8")) as {
            bomFormat: string;
            components: { name: string }[];
        };
        expect(sbom.bomFormat).toBe("CycloneDX");
        expect(sbom.components.map(c => c.name)).toContain("apk-tools");
    });

    await test.step("the project's Security tab lists the app, with its CRITICAL, and leads to its findings", async () => {
        await page.goto(`/projects/${app.projectId}/security/`);
        const tab = page.getByTestId("project-image-scan");
        await expect(tab.getByText("Beta")).toBeVisible();
        const row = tab.getByTestId("image-scan-app-row").filter({ hasText: OLD_ALPINE });
        await expect(row).toHaveCount(1);
        await row.getByRole("link", { name: app.name }).click();
        await expect(page).toHaveURL(/\/security\/$/);
        await expect(page.getByTestId("finding-row").filter({ hasText: CVE })).toBeVisible();
    });

    await test.step("the daily scan scans HivePaaS's own images: Traefik is listed, and opens its findings", async () => {
        await componentScanned(api, "traefik");
        await page.goto("/system/image-scanning/hivepaas/");
        const row = page.getByTestId("image-scan-component-row").filter({ hasText: "Traefik" });
        await expect(row).toContainText("traefik");
        await row.getByRole("link", { name: "Traefik" }).click();
        await expect(page.getByTestId("scanned-image")).toContainText("traefik");
        await expect(page.getByTestId("severity-critical")).toBeVisible();
    });

    await test.step("a project's ignore goes out with its spec, and an import brings it back", async () => {
        const name = e2eName("spec-ignore");
        cleanup(async () => {
            const made = await findProject(api, name);
            if (made) await deleteProject(api, made.id);
        });
        const project = await createProject(api, name);
        await createSettingAt(api, `projects/${project.id}/vulnerability-ignores`, {
            vulnId: CVE,
            package: "apk-tools",
            reason: "apk never runs in the container",
            expireAt: null,
            inheritable: true,
        });

        await page.goto(`/projects/${project.id}/operations/export/`);
        const downloading = page.waitForEvent("download");
        await page.getByRole("button", { name: "Export", exact: true }).click();
        const bundle = test.info().outputPath("bundle.tar.gz");
        await (await downloading).saveAs(bundle);
        await deleteProject(api, project.id);

        await page.goto("/operations/export/");
        const choosing = page.waitForEvent("filechooser");
        await page.getByRole("button", { name: "Choose file" }).click();
        await (await choosing).setFiles(bundle);
        await expect(page.getByRole("checkbox", { name: new RegExp(`^Import ${name} `) })).toBeChecked();
        await page.getByRole("button", { name: /^Import( and accept \d+ issues?)?$/ }).click();
        await expect(page.getByText("Imported", { exact: true })).toBeVisible({ timeout: 60_000 });

        const restored = await findProject(api, name);
        expect(restored?.id, "a project of its own again").not.toBe(project.id);
        const res = await api.get(`projects/${restored!.id}/vulnerability-ignores`);
        expect(res.ok(), await res.text()).toBe(true);
        const ignores = ((await res.json()) as { data: { vulnId: string; reason: string; setBy: string }[] }).data;
        expect(ignores).toContainEqual(
            expect.objectContaining({ vulnId: CVE, reason: "apk never runs in the container", setBy: "admin" }),
        );
    });
});

interface ComponentFinding {
    vulnId: string;
    fixedVersion?: string;
}

// componentWithFinding is a part of HivePaaS whose last scan found a fixable
// vulnerability no ignore sets aside - the ones its page lists by default - in
// one package, and that vulnerability.
async function componentWithFinding(api: APIRequestContext): Promise<{ key: string; vulnId: string }> {
    const res = await api.get("system/settings/image-scan/hivepaas");
    const { components } = ((await res.json()) as { data: { components: ScannedComponent[] } }).data;
    for (const { component } of components) {
        const detail = await api.get(`system/settings/image-scan/hivepaas/${component.key}`);
        const findings = ((await detail.json()) as { data: { scan: { findings: ComponentFinding[] } } }).data.scan
            .findings;
        // One of a single package: the dialog ignores it in that package.
        const fixable = findings.find(f => f.fixedVersion && findings.filter(o => o.vulnId === f.vulnId).length === 1);
        if (fixable) return { key: component.key, vulnId: fixable.vulnId };
    }
    throw new Error("no part of HivePaaS runs a fixable vulnerability to ignore");
}

interface GlobalIgnore {
    id: string;
    vulnId: string;
    inheritable: boolean;
}

async function globalIgnores(api: APIRequestContext): Promise<GlobalIgnore[]> {
    const res = await api.get("settings/vulnerability-ignores");
    expect(res.ok(), await res.text()).toBe(true);
    return ((await res.json()) as { data: GlobalIgnore[] }).data;
}

async function deleteGlobalIgnores(api: APIRequestContext, vulnId: string): Promise<void> {
    for (const ignore of (await globalIgnores(api)).filter(i => i.vulnId === vulnId)) {
        await api.delete(`settings/vulnerability-ignores/${ignore.id}`);
    }
}

test("image scanning: an ignore for HivePaaS only sets a finding of HivePaaS's images aside, and none of an app's; removed, it counts again", async ({
    page,
    api,
    cleanup,
}) => {
    cleanup(() => switchOff(api));
    const volume = await scanVolume(api, cleanup);
    await putSettings(api, true, volume.id);
    await componentScanned(api, "traefik");
    const part = await componentWithFinding(api);
    cleanup(() => deleteGlobalIgnores(api, part.vulnId));

    await test.step("ignored for HivePaaS only on a part's page, it leaves the findings for Ignored", async () => {
        await page.goto(`/system/image-scanning/hivepaas/${part.key}/`);
        await page
            .getByTestId("finding-row")
            .filter({ hasText: part.vulnId })
            .first()
            .getByRole("button", { name: "Ignore…" })
            .click();
        const dialog = page.getByRole("dialog");
        await expect(dialog.getByRole("combobox", { name: "Scope" })).toHaveText("HivePaaS only");
        await dialog.getByRole("textbox", { name: "Reason" }).fill("not reachable in HivePaaS's setup");
        await dialog.getByRole("button", { name: "Ignore", exact: true }).click();
        await expect(page.getByText("Vulnerability ignored")).toBeVisible();

        await expect(page.getByTestId("finding-row").filter({ hasText: part.vulnId })).toHaveCount(0);
        await page.getByRole("button", { name: /^Ignored \(\d+\)$/ }).click();
        const ignored = page.getByTestId("ignored-row").filter({ hasText: part.vulnId }).first();
        await expect(ignored).toContainText("HivePaaS only");
        await expect(ignored).toContainText("not reachable in HivePaaS's setup");

        const made = (await globalIgnores(api)).find(i => i.vulnId === part.vulnId);
        expect(made?.inheritable, "global, and not the apps'").toBe(false);
    });

    await test.step("an ignore for HivePaaS only leaves an app's finding counted", async () => {
        const app = await appIn(api, cleanup, "image-hp-only");
        await deployImage(api, app, OLD_ALPINE, "sleep 3600");
        await deployed(api, app);
        cleanup(() => deleteGlobalIgnores(api, CVE));
        await createSettingAt(api, "settings/vulnerability-ignores", {
            vulnId: CVE,
            reason: "HivePaaS's images only",
            expireAt: null,
            inheritable: false,
        });
        await page.goto(appPage(app, "security"));
        await criticalShown(page);
        await expect(page.getByTestId("finding-row").filter({ hasText: CVE })).toBeVisible();
        await expect(page.getByRole("button", { name: /^Ignored \(\d+\)$/ })).toHaveCount(0);
    });

    await test.step("removed, the part's finding counts again", async () => {
        await page.goto(`/system/image-scanning/hivepaas/${part.key}/`);
        await page.getByRole("button", { name: /^Ignored \(\d+\)$/ }).click();
        await page
            .getByTestId("ignored-row")
            .filter({ hasText: part.vulnId })
            .first()
            .getByRole("button", { name: "Remove" })
            .click();
        await page.getByRole("button", { name: "Remove", exact: true }).last().click();
        await expect(page.getByText("Ignore removed")).toBeVisible();
        await expect(page.getByTestId("finding-row").filter({ hasText: part.vulnId }).first()).toBeVisible();
    });
});

interface ClusterNode {
    id: string;
    labels: Record<string, string> | null;
    availability: string;
    updateVer: number;
}

// theNode is the e2e env's one node, with its labels.
async function theNode(api: APIRequestContext): Promise<ClusterNode> {
    const res = await api.get("cluster/nodes");
    expect(res.ok(), await res.text()).toBe(true);
    const nodes = ((await res.json()) as { data: ClusterNode[] }).data;
    expect(nodes, "the e2e env has one node").toHaveLength(1);
    // The list leaves the labels out: the node's own answer has them.
    const one = await api.get(`cluster/nodes/${nodes[0]!.id}`);
    expect(one.ok(), await one.text()).toBe(true);
    return ((await one.json()) as { data: ClusterNode }).data;
}

async function setNodeLabels(api: APIRequestContext, labels: Record<string, string>): Promise<void> {
    const node = await theNode(api);
    const res = await api.put(`cluster/nodes/${node.id}`, {
        data: { updateVer: node.updateVer, labels, availability: node.availability },
    });
    expect(res.ok(), await res.text()).toBe(true);
}

// labelTheNode gives the node a label of the test's, and takes it away after.
async function labelTheNode(api: APIRequestContext, cleanup: Cleanup, key: string): Promise<void> {
    const before = (await theNode(api)).labels ?? {};
    await setNodeLabels(api, { ...before, [key]: "yes" });
    cleanup(async () => {
        const { [key]: _, ...rest } = (await theNode(api)).labels ?? {};
        await setNodeLabels(api, rest);
    });
}

// pinnedVolume makes a volume pinned by a node label, or by nothing - one that
// claims every node reaches it - and deletes it once scanning is off.
async function pinnedVolume(
    api: APIRequestContext,
    cleanup: Cleanup,
    pin: { nodeLabel: string } | "shared",
): Promise<{ id: string; name: string }> {
    const name = e2eName(`trivy-${Date.now().toString(36)}`);
    const id = await created(api, "cluster/volumes", {
        name,
        driver: "local",
        inheritable: true,
        ...(pin === "shared" ? {} : pin),
    });
    cleanup(async () => {
        await switchOff(api);
        await deleteVolume(api, id);
    });
    return { id, name };
}

test("image scanning: a volume pinned by a node label runs scans on a node carrying it; one no node carries says so", async ({
    page,
    api,
    cleanup,
}) => {
    cleanup(() => switchOff(api));
    const key = `hp-e2e-scan-${Date.now().toString(36)}`;
    await labelTheNode(api, cleanup, key);
    const volume = await pinnedVolume(api, cleanup, { nodeLabel: `${key}=yes` });

    await test.step("chosen, the volume says scans run on a node carrying its label", async () => {
        await page.goto(SETTINGS);
        const enabled = page.getByRole("checkbox", { name: "Enabled" });
        if (!(await enabled.isChecked())) {
            await enabled.click();
        }
        await page
            .getByRole("group", { name: /^Data volume/ })
            .getByRole("combobox")
            .click();
        await page.getByRole("option", { name: volume.name }).click();
        await expect(page.getByTestId("image-scan-where")).toContainText(`a node carrying the label ${key}=yes`);
        await page.getByRole("button", { name: "Save" }).click();
        await expect(page.getByText("Image scanning settings updated")).toBeVisible();
        await expect(page.getByTestId("image-scanner-status").getByRole("group", { name: "Scans run" })).toContainText(
            `On a node carrying the label ${key}=yes`,
        );
    });

    const app = await appIn(api, cleanup, "image-label");
    await test.step("an app's image is scanned there", async () => {
        await deployImage(api, app, OLD_ALPINE, "sleep 3600");
        await deployed(api, app);
        await page.goto(appPage(app, "security"));
        await criticalShown(page);
        await expect(page.getByTestId("scanned-image")).toContainText(OLD_ALPINE);
    });

    await test.step("a label no node carries: the status says no scan can run, and a scan says why on the app", async () => {
        const nowhere = `${key}-nowhere=yes`;
        // Its cleanup deletes it too: a volume of a label no node carries is on
        // no node, and goes with nothing to remove.
        const unreachable = await pinnedVolume(api, cleanup, { nodeLabel: nowhere });
        await putSettings(api, true, unreachable.id);
        const said = `No ready node carries the label "${nowhere}"`;

        await page.goto(SETTINGS);
        await expect(page.getByTestId("image-scan-volume-problem")).toContainText(`${said}: no scan can run.`);

        await page.goto(appPage(app, "security"));
        await page.getByRole("button", { name: "Scan now" }).click();
        await expect(async () => {
            await page.reload();
            await expect(page.getByText(said)).toBeVisible({ timeout: 5_000 });
        }).toPass(POLL);
        // What it ran last stays.
        await expect(page.getByTestId("finding-row").filter({ hasText: CVE })).toBeVisible();
    });
});

test("image scanning: a shared volume runs each scan on the node running its image", async ({ page, api, cleanup }) => {
    cleanup(() => switchOff(api));
    const volume = await pinnedVolume(api, cleanup, "shared");
    await putSettings(api, true, volume.id);

    await page.goto(SETTINGS);
    await expect(page.getByTestId("image-scanner-status").getByRole("group", { name: "Scans run" })).toContainText(
        "On the node running each image: the volume is shared",
    );

    const app = await appIn(api, cleanup, "image-shared");
    await deployImage(api, app, OLD_ALPINE, "sleep 3600");
    await deployed(api, app);
    await page.goto(appPage(app, "security"));
    await criticalShown(page);
    await expect(page.getByTestId("scanned-image")).toContainText(OLD_ALPINE);
});

interface ReportAttachment {
    title: string;
    text: string;
}

// runDaily runs the daily scan now, and waits for it to end.
async function runDaily(api: APIRequestContext): Promise<void> {
    const run = await api.post("system/settings/image-scan/exec", { data: {} });
    expect(run.ok(), await run.text()).toBe(true);
    const id = ((await run.json()) as { data: { task: { id: string } } }).data.task.id;
    await expect
        .poll(async () => {
            const res = await api.get(`system/tasks/${id}`);
            return ((await res.json()) as { data: { status: string } }).data.status;
        }, POLL)
        .toMatch(/^(done|failed|canceled)$/);
}

test("image scanning: the daily scan tells its notification what apps run that is new, once", async ({
    page,
    api,
    cleanup,
}) => {
    cleanup(() => switchOff(api));
    const volume = await scanVolume(api, cleanup);

    // An echo app stands in for Slack, through a Slack and a notification of
    // the installation's: the daily scan's are global.
    const project = await createProject(api, e2eName("scan-report"));
    cleanup(() => deleteProject(api, project.id));
    const { hook, domain } = await echoHook(page, api, project, "scan-report-hook");
    const slackName = e2eName("scan-slack");
    const notificationName = e2eName("scan-report");
    cleanup(() => deleteSettingsNamed(api, "im-services", slackName));
    const slack = await created(api, "settings/im-services", {
        name: slackName,
        kind: "slack",
        slack: { webhook: `http://${domain}/slack` },
    });
    cleanup(() => deleteSettingsNamed(api, "notifications", notificationName));
    const notification = await created(api, "settings/notifications", {
        name: notificationName,
        viaSlack: { enabled: true, useDefault: false, webhook: { id: slack } },
    });
    // Put back as it was before the notification goes: it refers to it.
    const before = (await imageScanSettings(api)).notification;
    cleanup(() => putSettings(api, false, undefined, before));
    await putSettings(api, true, volume.id, {
        success: { id: notification },
        successUseDefault: false,
        failureUseDefault: false,
    });

    const app = await appIn(api, cleanup, "image-report");
    // A report's line names the app by its project, env and name.
    const line = `${e2eName("image-report")} / `;
    await deployImage(api, app, OLD_ALPINE, "sleep 3600");
    await deployed(api, app);
    await page.goto(appPage(app, "security"));
    await criticalShown(page);

    const reports = async () =>
        ((await slackMessages(page, hook)) as unknown as ReportAttachment[]).filter(m =>
            m.title.includes("new fixable vulnerabilities"),
        );

    await test.step("the first daily scan after the app's scan reports it", async () => {
        await runDaily(api);
        let found: ReportAttachment | undefined;
        await expect(async () => {
            found = (await reports()).find(m => m.text.includes(line));
            expect(found).toBeDefined();
        }).toPass(POLL);
        expect(found?.text).toContain("1 CRITICAL");
        expect(found?.title).toMatch(/\d+ apps? runs? new fixable vulnerabilities/);
    });

    await test.step("the next one has nothing new to tell of it", async () => {
        const told = (await reports()).filter(m => m.text.includes(line)).length;
        await runDaily(api);
        // The report goes after the run's end: give it the time it would take.
        await page.waitForTimeout(5_000);
        expect((await reports()).filter(m => m.text.includes(line))).toHaveLength(told);
    });
});
