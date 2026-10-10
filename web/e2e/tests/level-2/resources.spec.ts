import type { APIRequestContext, Page } from "@playwright/test";

import { type App, appPath, deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectInstances, expectLogs, expectPrinted } from "../../support/apps";
import { type Cleanup, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 240_000 });

// The container prints the limits its cgroup has - "max" for none - then asks
// for more memory than 64 MB at once: dd's exit code says whether it got it, or
// was killed for it.
const PRINT_LIMITS =
    "sh -c 'echo memory.max=$(cat /sys/fs/cgroup/memory.max); echo swap.max=$(cat /sys/fs/cgroup/memory.swap.max); " +
    "echo cpu.max=$(cat /sys/fs/cgroup/cpu.max); " +
    "echo pids.max=$(cat /sys/fs/cgroup/pids.max); dd if=/dev/zero of=/dev/null bs=128M count=1 2>/dev/null; " +
    "echo dd=$?; exec sleep 3600'";

test("resource limits saved apply to the container: CPU, memory, swap, processes; memory over them is killed", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "limits");
    await deployImage(api, app, BUSYBOX, PRINT_LIMITS);
    await deployed(api, app);
    // Unlimited until limits are set.
    await expectLogs(page, app, "memory.max=max");
    await expectLogs(page, app, "dd=0");

    await page.goto(appPage(app, "resources"));
    const limits = page.getByText("Resource Limit", { exact: true }).locator("xpath=..");
    await limits.getByRole("group", { name: "CPUs" }).getByRole("textbox").fill("0.5");
    await limits.getByRole("group", { name: "Memory" }).getByRole("textbox").fill("64mb");
    await limits.getByRole("group", { name: "Pids" }).getByRole("textbox").fill("64");
    // No swap to spill into: as much as memory, unless told.
    await page.getByRole("group", { name: "Swap Memory" }).getByRole("textbox").fill("0");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Resource settings updated")).toBeVisible();

    // Saved, they apply at once: the next container runs within them.
    await expectLogs(page, app, "memory.max=67108864");
    await expectLogs(page, app, "swap.max=0");
    await expectLogs(page, app, "cpu.max=50000 100000");
    await expectLogs(page, app, "pids.max=64");
    await expectLogs(page, app, "dd=137");
});

// saying is a container that says what it has, key=value a line, every few
// seconds - with the time, which tells the newest container's from one gone
// that the log still shows.
const saying = (values: Record<string, string>) =>
    `sh -c 'while true; do ${Object.entries(values)
        .map(([key, value]) => `echo ${key}=${value}@$(date +%s)`)
        .join("; ")}; sleep 3; done'`;

// appSaying is an app of its own running a container saying so.
async function appSaying(
    page: Page,
    api: APIRequestContext,
    cleanup: Cleanup,
    label: string,
    command: string,
): Promise<App> {
    const app = await appIn(api, cleanup, label);
    await deployImage(api, app, BUSYBOX, command);
    await deployed(api, app);
    await page.goto(appPage(app, "logs"));
    return app;
}

// saveResources saves the app's Resources page as it is filled.
async function saveResources(page: Page): Promise<void> {
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Resource settings updated")).toBeVisible();
}

const field = (page: Page, name: string) => page.getByRole("group", { name, exact: true });
const section = (page: Page, label: string) => page.getByText(label, { exact: true }).locator("xpath=..");

// A reservation is what swarm sets aside for the container on the node it
// places it on: one no node can meet places nothing, and the app runs again
// once it is lowered. It is the scheduler's, not the container's: the
// container itself is not limited by it.
test("a reservation no node can meet runs nothing, and lowered, the app runs again", async ({ page, api, cleanup }) => {
    const app = await appSaying(page, api, cleanup, "reserve", saying({ running: "yes" }));
    await expectPrinted(page, app, "running", "yes");

    // More CPUs than the node has, and memory it has.
    await page.goto(appPage(app, "resources"));
    const reservation = section(page, "Resource Reservation");
    await reservation.getByRole("group", { name: "CPUs" }).getByRole("textbox").fill("64");
    await reservation.getByRole("group", { name: "Memory" }).getByRole("textbox").fill("32mb");
    await saveResources(page);
    await expectInstances(page, app, "0/1");
    await page.goto(appPage(app, "resources"));
    await expect(reservation.getByRole("group", { name: "CPUs" }).getByRole("textbox")).toHaveValue("64");
    await expect(reservation.getByRole("group", { name: "Memory" }).getByRole("textbox")).toHaveValue("32mb");

    await reservation.getByRole("group", { name: "CPUs" }).getByRole("textbox").fill("");
    await saveResources(page);
    await expectInstances(page, app, "1/1");
});

// Shared memory and swap as saved; cleared, the container has what it had
// before them. Swappiness - which a cgroup v2 container does not have - is
// kept as saved.
test("shared memory and swap apply as saved, and cleared, are as before; swappiness is kept", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appSaying(
        page,
        api,
        cleanup,
        "memory",
        saying({
            shm: '$(df -m /dev/shm | tail -1 | tr -s " " | cut -d" " -f2)',
            swap: "$(cat /sys/fs/cgroup/memory.swap.max)",
        }),
    );
    await expectPrinted(page, app, "shm", "64");

    // A memory limit, and the swap it has without one said.
    await page.goto(appPage(app, "resources"));
    await section(page, "Resource Limit").getByRole("group", { name: "Memory" }).getByRole("textbox").fill("64mb");
    await saveResources(page);
    const defaultSwap = await expectPrinted(page, app, "swap", printed => printed !== "" && printed !== "max");

    await page.goto(appPage(app, "resources"));
    await field(page, "Swap Memory").getByRole("textbox").fill("32mb");
    await field(page, "Shm Size").getByRole("textbox").fill("128mb");
    await field(page, "Swappiness").getByRole("textbox").fill("10");
    await field(page, "Swappiness").getByRole("textbox").blur();
    await saveResources(page);
    await expectPrinted(page, app, "shm", "128");
    await expectPrinted(page, app, "swap", String(32 * 1024 * 1024));
    await page.goto(appPage(app, "resources"));
    await expect(field(page, "Swappiness").getByRole("textbox")).toHaveValue("10");

    await field(page, "Swap Memory").getByRole("textbox").fill("");
    await field(page, "Shm Size").getByRole("textbox").fill("");
    await saveResources(page);
    await expectPrinted(page, app, "shm", "64");
    await expectPrinted(page, app, "swap", defaultSwap);
});

// Ulimits added apply to the container's processes, soft and hard; one
// removed is as before.
test("ulimits apply to the container, soft and hard, and one removed is as before", async ({ page, api, cleanup }) => {
    const app = await appSaying(
        page,
        api,
        cleanup,
        "ulimits",
        saying({ nofile: "$(ulimit -Sn)/$(ulimit -Hn)", nproc: "$(ulimit -Su)/$(ulimit -Hu)" }),
    );
    const nofile = await expectPrinted(page, app, "nofile", printed => printed.includes("/"));
    await expectPrinted(page, app, "nproc", "unlimited/unlimited");

    await page.goto(appPage(app, "resources"));
    const ulimits = field(page, "Ulimits");
    for (const [name, soft, hard] of [
        ["nofile", "1000", "2000"],
        ["nproc", "300", "400"],
    ] as const) {
        await ulimits.getByPlaceholder("nofile").fill(name);
        await ulimits.getByPlaceholder("nofile").press("Escape");
        await ulimits.getByPlaceholder("1024").nth(0).fill(soft);
        await ulimits.getByPlaceholder("1024").nth(1).fill(hard);
        await ulimits.getByPlaceholder("1024").nth(1).blur();
        await ulimits.getByRole("button", { name: "Add" }).click();
    }
    await saveResources(page);
    await expectPrinted(page, app, "nofile", "1000/2000");
    await expectPrinted(page, app, "nproc", "300/400");

    await page.goto(appPage(app, "resources"));
    await ulimits
        .getByText("nproc", { exact: true })
        .locator("xpath=../..")
        .getByRole("button", { name: "Remove item" })
        .click();
    await page.getByRole("dialog").getByRole("button", { name: "Remove" }).click();
    await saveResources(page);
    await expectPrinted(page, app, "nproc", "unlimited/unlimited");
    await expectPrinted(page, app, "nofile", "1000/2000");
    expect(nofile, "a limit of its own until then").not.toBe("1000/2000");
});

// Capabilities added and dropped, the out-of-memory score and a sysctl are the
// container's: it may set an interface's MTU, may not change a file's owner,
// and reads the score and the kernel parameter as saved. Cleared, it is as it
// was.
test("capabilities added and dropped, the out-of-memory score and a sysctl apply to the container, and cleared, are as before", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appSaying(
        page,
        api,
        cleanup,
        "capabilities",
        saying({
            mtu: "$(ip link set lo mtu 1400 2>/dev/null && echo set || echo refused)",
            chown: "$(chown 1 /tmp 2>/dev/null && echo done || echo refused)",
            oom: "$(cat /proc/self/oom_score_adj)",
            somaxconn: "$(cat /proc/sys/net/core/somaxconn)",
        }),
    );
    const before = { mtu: "refused", chown: "done", oom: "0", somaxconn: "4096" };
    for (const [key, value] of Object.entries(before)) {
        await expectPrinted(page, app, key, value);
    }

    await page.goto(appPage(app, "resources"));
    await field(page, "Capabilities Add").getByRole("textbox").fill("NET_ADMIN");
    await field(page, "Capabilities Drop").getByRole("textbox").fill("CHOWN");
    await field(page, "Out-of-Mem Score Adjustment").getByRole("textbox").fill("500");
    await field(page, "Out-of-Mem Score Adjustment").getByRole("textbox").blur();
    const sysctls = field(page, "Sysctls");
    await sysctls.getByPlaceholder("net.core.somaxconn").fill("net.core.somaxconn");
    await sysctls.getByPlaceholder("1024").fill("777");
    await sysctls.getByRole("button", { name: "Add" }).click();
    await saveResources(page);
    for (const [key, value] of Object.entries({ mtu: "set", chown: "refused", oom: "500", somaxconn: "777" })) {
        await expectPrinted(page, app, key, value);
    }

    await page.goto(appPage(app, "resources"));
    await field(page, "Capabilities Add").getByRole("textbox").fill("");
    await field(page, "Capabilities Drop").getByRole("textbox").fill("");
    await field(page, "Out-of-Mem Score Adjustment").getByRole("textbox").fill("");
    await field(page, "Out-of-Mem Score Adjustment").getByRole("textbox").blur();
    await sysctls.getByRole("button", { name: "Remove item" }).click();
    await page.getByRole("dialog").getByRole("button", { name: "Remove" }).click();
    await saveResources(page);
    for (const [key, value] of Object.entries(before)) {
        await expectPrinted(page, app, key, value);
    }
});

// The page is saved against what it shows of the service, not against the
// service's own version: loaded before swarm marked the deployment's rolling
// update complete - which writes the service again - it saves; another change
// to what it shows, made while it was open, is refused.
test("a page loaded before a deployment settled saves; one whose settings changed meanwhile is refused", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "settled");
    await deployImage(api, app, BUSYBOX, "sh -c 'exec sleep 3600'");
    await deployed(api, app);
    const versionNow = async () => {
        const res = await api.get(`${appPath(app)}/resource-settings`);
        return ((await res.json()) as { data: { updateVer: number; limits: unknown } }).data;
    };

    // Loaded at once, and saved once swarm has written the service again.
    await page.goto(appPage(app, "resources"));
    const loaded = (await versionNow()).updateVer;
    await page.waitForTimeout(12_000);
    expect((await versionNow()).updateVer, "the page's version stays").toBe(loaded);
    await section(page, "Resource Limit").getByRole("group", { name: "Pids" }).getByRole("textbox").fill("64");
    await saveResources(page);

    // Open again, and the limits changed by someone else meanwhile.
    await page.goto(appPage(app, "resources"));
    await expect(section(page, "Resource Limit").getByRole("group", { name: "Pids" }).getByRole("textbox")).toHaveValue(
        "64",
    );
    const current = await versionNow();
    const res = await api.put(`${appPath(app)}/resource-settings`, {
        data: { ...current, limits: { pids: 32 } },
    });
    expect(res.ok(), `changing the limits: ${res.status()} ${await res.text()}`).toBe(true);
    await section(page, "Resource Limit").getByRole("group", { name: "Pids" }).getByRole("textbox").fill("128");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Mismatching update version").first()).toBeVisible();
    expect(((await versionNow()).limits as { pids?: number }).pids, "the other change stands").toBe(32);
});
