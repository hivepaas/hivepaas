import type { Locator, Page } from "@playwright/test";

import { type App, createApp, createVolume, deleteVolume, deployImage } from "../../support/api";
import {
    BUSYBOX,
    appIn,
    appPage,
    deployed,
    expectLogs,
    fillMount,
    mountRow,
    mountVolume,
    restart,
} from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 240_000 });

// The container writes a mark on its first run and prints it on the next: what
// it wrote to the volume is still there for the container that replaces it.
const KEEPS_A_MARK =
    'sh -c \'if [ -f /data/mark ]; then echo "kept-$(cat /data/mark)"; ' +
    'else echo "$E2E_MARK" > /data/mark; echo wrote; fi; exec sleep 3600\'';

test("a volume mounted in an app keeps what it wrote for the next container", async ({ page, api, cleanup }) => {
    // Made before the project, so that it is removed after it: the steps run
    // the last added first, and the app mounts the volume.
    const volume = e2eName("kept-data");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, "volume");

    // Listed by its name: docker knows it by its id.
    await mountVolume(page, app, volume, "/data");

    await deployImage(api, app, BUSYBOX, KEEPS_A_MARK.replace("$E2E_MARK", volume));
    await deployed(api, app);
    await expectLogs(page, app, "wrote");

    await restart(page, app);
    await expectLogs(page, app, `kept-${volume}`);
});

// subpaths: two directories of the volume, and the first again, read only.
test("a subpath is a directory of its own, mounted twice is the same one, and read only refuses writes", async ({
    page,
    api,
    cleanup,
}) => {
    const volume = e2eName("subpaths");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, "subpaths");

    await mountVolume(page, app, volume, "/one", { subpath: "one" });
    await mountVolume(page, app, volume, "/two", { subpath: "two" });
    await mountVolume(page, app, volume, "/again", { subpath: "one", readOnly: true });

    await deployImage(
        api,
        app,
        BUSYBOX,
        "sh -c 'echo x > /one/f; touch /again/g 2>/dev/null && w=writable || w=read-only; " +
            'echo "two=[$(ls /two)] again=[$(ls /again)] $w"; exec sleep 3600\'',
    );
    await deployed(api, app);
    // What /one wrote is not in /two, and is in /again - which takes no write.
    await expectLogs(page, app, "two=[] again=[f] read-only");
});

// Data of: one app reads another's directory - only reads, until allowed to
// write - and the other's screen says who reads its data.
test("an app given another's data reads it, writes only when allowed, and the owner sees who reads it", async ({
    page,
    api,
    cleanup,
}) => {
    const volume = e2eName("shared");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    const owner = await appIn(api, cleanup, "storage-owner");
    await mountVolume(page, owner, volume, "/data");
    await deployImage(api, owner, BUSYBOX, "sh -c 'echo owned > /data/owned; echo owner-ready; exec sleep 3600'");
    await deployed(api, owner);
    await expectLogs(page, owner, "owner-ready");

    const reader = await createApp(api, { id: owner.projectId }, "reader");
    await mountVolume(page, reader, volume, "/shared", { dataOf: owner.name });
    await deployImage(
        api,
        reader,
        BUSYBOX,
        'sh -c \'s=$(ls /shared | tr "\\n" " "); touch /shared/by-reader 2>/dev/null && w=can-write || w=cannot-write; ' +
            'echo "sees=[$s] $w"; exec sleep 3600\'',
    );
    await deployed(api, reader);
    await expectLogs(page, reader, "sees=[owned ] cannot-write");

    // The owner's screen names the app that reads its data.
    await page.goto(appPage(owner, "persistent-storage"));
    const readers = page.getByText("Apps reading this app's storage", { exact: true }).locator("xpath=..");
    await expect(readers).toContainText(/reader\s*reads\s*everything\s*at \/shared/);

    // Allowed to write, it writes.
    await page.goto(appPage(reader, "persistent-storage"));
    await mountRow(page, "/shared").getByRole("button", { name: "Edit storage mount" }).click();
    await page.getByRole("group", { name: "Allow writing", exact: true }).getByRole("checkbox").check();
    await page.getByRole("button", { name: "Update" }).click();
    await expect(page.getByText("Storage mount updated")).toBeVisible();
    await expectLogs(page, reader, "sees=[owned ] can-write");
    await page.goto(appPage(owner, "persistent-storage"));
    await expect(readers).toContainText(/reader\s*reads and changes\s*everything/);
});

// Each container marks what it finds unmarked, with the time, and says what
// /data and /store hold: whose data a container reads is in its log.
const MARKS =
    "sh -c 'for d in /data /store; do [ -d $d ] && [ ! -f $d/mark ] && echo made-$RANDOM$RANDOM > $d/mark; done; " +
    'echo "data=$(cat /data/mark 2>/dev/null || echo none) store=$(cat /store/mark 2>/dev/null || echo none)"; ' +
    "exec sleep 3600'";

// A mount moved keeps its data; removed, the data stays on the volume; mounted
// again, the form says the directory has data, and keeps it or deletes it.
test("a mount moved keeps its data; removed and mounted again, its data is kept or deleted, as asked", async ({
    page,
    api,
    cleanup,
}) => {
    const volume = e2eName("in-use");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, "in-use");
    await mountVolume(page, app, volume, "/data");
    await deployImage(api, app, BUSYBOX, MARKS);
    await deployed(api, app);
    const first = /data=(made-\d+) store=none/.exec(await expectLogs(page, app, /data=made-\d+ store=none/))![1]!;

    // Moved to /store: the same directory, its mark with it.
    await page.goto(appPage(app, "persistent-storage"));
    await mountRow(page, "/data").getByRole("button", { name: "Edit storage mount" }).click();
    await page.getByRole("group", { name: "Target *" }).getByRole("textbox").fill("/store");
    await saveMount(page, "Update", "Save anyway");
    await expectLogs(page, app, `data=none store=${first}`);

    // Removed: the container has neither.
    await removeMount(page, app, "/store");
    await expectLogs(page, app, "data=none store=none");

    // Mounted again: its data is there - kept.
    await page.goto(appPage(app, "persistent-storage"));
    await page.getByRole("button", { name: "New Storage Mount" }).click();
    await fillMount(page, volume, "/data");
    await saveMount(page, "Save", "Save anyway");
    await expectLogs(page, app, `data=${first} store=none`);

    // Removed, and mounted again with its data deleted: a new mark.
    await removeMount(page, app, "/data");
    await expectLogs(page, app, "data=none store=none");
    await page.goto(appPage(app, "persistent-storage"));
    await page.getByRole("button", { name: "New Storage Mount" }).click();
    await fillMount(page, volume, "/data");
    await saveMount(page, "Save", "Delete that data and save");
    // A mark of its own: the old one was deleted. (A container being replaced may
    // still be the one the log shows: the new mark is waited for.)
    await expectLogs(page, app, new RegExp(`data=made-(?!${first.slice("made-".length)} )\\d+ store=none`));
});

// Reset permissions: given to a user, the files are that user's, their modes
// kept; opened to every user, anyone reads and writes them.
test("the files of a mount are given to a user, then opened to every user", async ({ page, api, cleanup }) => {
    const volume = e2eName("perms");
    const volumeId = await createVolume(api, volume);
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, "perms");
    await mountVolume(page, app, volume, "/data");
    // Written as root, once; each container says whose it is and its mode.
    await deployImage(
        api,
        app,
        BUSYBOX,
        'sh -c \'[ -f /data/f ] || echo x > /data/f; echo "f=$(stat -c "%u:%g %a" /data/f)"; exec sleep 3600\'',
    );
    await deployed(api, app);
    await expectLogs(page, app, "f=0:0 644");

    await resetPermissions(page, app, "/data", async dialog => {
        await dialog.getByRole("button", { name: /^Give to a user/ }).click();
        await dialog.getByLabel("User ID").fill("1000");
        await dialog.getByLabel("Group ID").fill("1000");
    });
    await restart(page, app);
    await expectLogs(page, app, "f=1000:1000 644");

    await resetPermissions(page, app, "/data", async dialog => {
        await dialog.getByRole("button", { name: /^Open to every user/ }).click();
    });
    await restart(page, app);
    await expectLogs(page, app, "f=1000:1000 666");
});

// A volume that is a directory of the node is mounted as that directory: its
// data is the node's, kept for the next container.
test("a volume bound to a directory of the node keeps what the app wrote there", async ({ page, api, cleanup }) => {
    const volume = e2eName("bound");
    const volumeId = await createVolume(api, volume, `/hp/${volume}`);
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, "bound");

    await mountVolume(page, app, volume, "/data");
    await expect(mountRow(page, "/data")).toContainText("bind");

    await deployImage(api, app, BUSYBOX, KEEPS_A_MARK.replace("$E2E_MARK", volume));
    await deployed(api, app);
    await expectLogs(page, app, "wrote");
    await restart(page, app);
    await expectLogs(page, app, `kept-${volume}`);
});

// saveMount saves the form - Save for a new mount, Update for one edited; the
// directory it reaches has data, and the dialog that says so is answered.
async function saveMount(
    page: Page,
    button: "Save" | "Update",
    answer: "Save anyway" | "Delete that data and save",
): Promise<void> {
    await page.getByRole("button", { name: button }).click();
    const dialog = page.getByRole("dialog", { name: "This storage already has data" });
    await dialog.getByRole("button", { name: answer }).click();
    await expect(page.getByText(/Storage mount (added|updated)/)).toBeVisible();
}

// removeMount removes the mount at the target from Persistent Storage.
async function removeMount(page: Page, app: App, target: string): Promise<void> {
    await page.goto(appPage(app, "persistent-storage"));
    await mountRow(page, target).getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: "Remove" }).click();
    await page.getByRole("dialog", { name: "Remove Storage Mount" }).getByRole("button", { name: "Remove" }).click();
    await expect(mountRow(page, target)).toHaveCount(0);
}

// resetPermissions resets the permissions of the mount's files, as `choose`
// picks in the dialog.
async function resetPermissions(
    page: Page,
    app: App,
    target: string,
    choose: (dialog: Locator) => Promise<void>,
): Promise<void> {
    await page.goto(appPage(app, "persistent-storage"));
    await mountRow(page, target).getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: "Reset permissions" }).click();
    const dialog = page.getByRole("dialog", { name: "Reset permissions" });
    await choose(dialog);
    await dialog.getByRole("button", { name: "Reset permissions" }).click();
    await expect(page.getByText(`Permissions reset for ${target}`)).toBeVisible();
}
