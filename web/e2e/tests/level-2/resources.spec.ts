import { deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs } from "../../support/apps";
import { expect, test } from "../../support/fixtures";

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
