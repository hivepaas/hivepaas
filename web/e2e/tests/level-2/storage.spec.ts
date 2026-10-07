import { createVolume, deleteVolume, deployImage } from "../../support/api";
import { BUSYBOX, appIn, deployed, expectLogs, mountVolume, restart } from "../../support/apps";
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
