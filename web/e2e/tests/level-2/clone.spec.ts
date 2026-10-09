import type { APIRequestContext, Page } from "@playwright/test";

import {
    type App,
    createVolume,
    deleteVolume,
    deployImage,
    findAppNamed,
    setRuntimeEnvVars,
    storageMounts,
} from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs, mountVolume, redeploy, restart } from "../../support/apps";
import { type Cleanup, e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

test("an app cloned in its environment runs, with its image and variables, and deploys", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "clone");
    const greeting = e2eName("greeting");
    await setRuntimeEnvVars(api, app, [{ key: "E2E_GREETING", value: greeting, isLiteral: false }]);
    await deployImage(api, app, BUSYBOX, "sh -c 'echo \"hello-$E2E_GREETING\"; exec sleep 3600'");
    await deployed(api, app);

    await page.goto(appPage(app, "app-clone"));
    await page.getByRole("group", { name: "Target Name" }).getByRole("textbox").fill("web-clone");
    await page.getByRole("button", { name: "Save and Clone" }).click();
    await expect(page.getByText("App clone started")).toBeVisible();

    // The clone goes with the project, as the app does.
    await expect.poll(() => findAppNamed(api, app, "web-clone"), { timeout: 120_000 }).toBeDefined();
    const clone = (await findAppNamed(api, app, "web-clone"))!;
    // A clone is made from the app's service, not deployed: it lists no
    // deployment, and its log is what says it runs.
    await expectLogs(page, clone, `hello-${greeting}`);

    // Its deployment settings came along: it deploys on its own.
    await redeploy(page, api, clone);
});

// Each of the app's containers counts its start in /data, signs it with its
// host name - the app's key - and says what /data holds: which app's data a
// container reads is in its log.
const COUNTS_ITS_STARTS =
    "sh -c 'mkdir -p /data; n=$(( $(cat /data/starts 2>/dev/null || echo 0) + 1 )); echo $n >/data/starts; " +
    'touch /data/by-$HOSTNAME; echo "starts=$n files=$(ls /data | tr "\\n" ",")"; exec sleep 3600\'';

// appWithData is an app with a volume mounted at /data, run once: its data is
// starts=1 and by-web.
async function appWithData(page: Page, api: APIRequestContext, cleanup: Cleanup, label: string): Promise<App> {
    // Made before the project, to be removed after it: the app mounts it.
    const volumeId = await createVolume(api, e2eName(`${label}-vol`));
    cleanup(() => deleteVolume(api, volumeId));
    const app = await appIn(api, cleanup, label);
    await mountVolume(page, app, e2eName(`${label}-vol`), "/data");
    await deployImage(api, app, BUSYBOX, COUNTS_ITS_STARTS);
    await deployed(api, app);
    await expectLogs(page, app, "starts=1 files=by-web,starts,");
    return app;
}

// cloneAs clones the app from its Clone page into the env, under a name, with
// the page's volume options set as given - each left as the page has it when
// not given.
async function cloneAs(
    page: Page,
    api: APIRequestContext,
    app: App,
    name: string,
    options: { volumes?: boolean; volumeData?: boolean; stopSource?: boolean },
): Promise<App> {
    await page.goto(appPage(app, "app-clone"));
    await page.getByRole("group", { name: "Target Name" }).getByRole("textbox").fill(name);
    const volumes = page.getByText("Clone Volumes", { exact: true }).locator("xpath=..");
    if (options.volumes !== undefined) {
        await volumes
            .getByRole("group", { name: "Enabled", exact: true })
            .getByRole("checkbox")
            .setChecked(options.volumes);
    }
    if (options.volumeData !== undefined) {
        await page
            .getByRole("group", { name: "Clone Volume Data" })
            .getByRole("checkbox")
            .setChecked(options.volumeData);
    }
    if (options.stopSource !== undefined) {
        await page
            .getByRole("group", { name: "Stop Source App Before Clone" })
            .getByRole("checkbox")
            .setChecked(options.stopSource);
    }
    await page.getByRole("button", { name: "Save and Clone" }).click();
    await expect(page.getByText("App clone started")).toBeVisible();
    await expect.poll(() => findAppNamed(api, app, name), { timeout: 120_000 }).toBeDefined();
    return (await findAppNamed(api, app, name))!;
}

// The clone's volumes are copies: what the source wrote is in the clone's, and
// what each writes after is its own. Stopped for the copy, the source is
// started again on its own.
test("an app cloned with its volumes' data runs on a copy, and the source, stopped for it, runs again", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appWithData(page, api, cleanup, "clone-data");

    const clone = await cloneAs(page, api, app, "web-clone", { volumes: true, volumeData: true, stopSource: true });

    // The copy holds the source's first start, and the clone's own.
    await expectLogs(page, clone, "starts=2 files=by-web,by-web-clone,starts,");
    // The source was stopped and started again: a second start, on its own data.
    await expectLogs(page, app, "starts=2 files=by-web,starts,");
});

// The clone's volumes made without the data: its directories are its own, and
// empty - its container starts on them.
test("an app cloned with its volumes but not their data starts on empty directories of its own", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appWithData(page, api, cleanup, "clone-empty");

    const clone = await cloneAs(page, api, app, "web-clone", { volumes: true, volumeData: false });

    await expectLogs(page, clone, "starts=1 files=by-web-clone,starts,");
    expect((await storageMounts(api, clone)).map(m => m.target)).toEqual(["/data"]);
    await restart(page, app);
    await expectLogs(page, app, "starts=2 files=by-web,starts,");
});

// The volumes not cloned, the clone has none: it never reaches the source's
// data, and the source reads nothing of the clone's.
test("an app cloned without its volumes has none, and leaves the source's data alone", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appWithData(page, api, cleanup, "clone-none");

    const clone = await cloneAs(page, api, app, "web-clone", { volumes: false });

    await expectLogs(page, clone, "starts=1 files=by-web-clone,starts,");
    expect(await storageMounts(api, clone), "the clone mounts nothing").toEqual([]);
    await page.goto(appPage(clone, "persistent-storage"));
    await expect(page.getByRole("row", { name: /\/data/ })).toHaveCount(0);

    await restart(page, app);
    await expectLogs(page, app, "starts=2 files=by-web,starts,");
});
