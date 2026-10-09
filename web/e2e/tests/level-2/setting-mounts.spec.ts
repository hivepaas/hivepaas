import type { APIRequestContext, Page } from "@playwright/test";

import { type App, appPath, createSettingAt, deployImage, latestDeployment } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed, expectLogs, expectPrinted } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

// The image a basic auth's htpasswd is checked with: Apache's own htpasswd
// reads the file.
const HTTPD = "httpd:2.4-alpine";

interface FileMount {
    part: string;
    path: string;
    mode?: string;
}

// newSettingMount saves a setting mount from the app's Setting Mounts: the
// files of the parts given, from a setting of the type picked.
async function newSettingMount(
    page: Page,
    app: App,
    mount: { name: string; from: string; setting: string; files: FileMount[] },
): Promise<void> {
    await page.goto(appPage(app, "setting-mounts"));
    await page.getByRole("button", { name: "New Setting Mount" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill(mount.name);
    await page.getByRole("combobox", { name: "Mount From" }).click();
    await page.getByRole("option", { name: mount.from, exact: true }).click();
    await page.getByRole("group", { name: "Setting *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: mount.setting, exact: true }).click();
    for (const file of mount.files) {
        await page.getByRole("checkbox", { name: file.part, exact: true }).check();
        await page.getByRole("textbox", { name: `${file.part} path`, exact: true }).fill(file.path);
        if (file.mode) {
            await page.getByRole("textbox", { name: `${file.part} mode`, exact: true }).fill(file.mode);
        }
    }
    await page.getByRole("button", { name: "Save" }).click();
}

// mountRow is the setting mount's row in the app's Setting Mounts.
async function mountRow(page: Page, app: App, name: string) {
    await page.goto(appPage(app, "setting-mounts"));
    return page.getByRole("row", { name: new RegExp(name) });
}

// switchMount turns a setting mount off or on from its row's menu.
async function switchMount(page: Page, app: App, name: string, action: "Disable" | "Enable"): Promise<void> {
    const row = await mountRow(page, app, name);
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: action, exact: true }).click();
    await expect(page.getByText(`Setting mount ${action.toLowerCase()}d`)).toBeVisible();
}

test("a file mounted from a config file follows it without a deploy; turned off it is gone, on again it is back", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "mount-follow");
    const before = e2eName("before");
    const after = e2eName("after");
    await createSettingAt(api, `${appPath(app)}/config-files`, { name: "E2E_CONF", content: before });

    await newSettingMount(page, app, {
        name: "e2e-conf",
        from: "Config file",
        setting: "E2E_CONF",
        files: [{ part: "content", path: "/etc/e2e.conf" }],
    });
    await expect(page.getByRole("row", { name: /e2e-conf/ })).toBeVisible();

    await deployImage(
        api,
        app,
        BUSYBOX,
        `sh -c 'while :; do echo "conf=$(cat /etc/e2e.conf 2>/dev/null || echo missing)"; sleep 2; done'`,
    );
    await deployed(api, app);
    const deployment = await latestDeployment(api, app);
    await expectPrinted(page, app, "conf", before);
    await expect(await mountRow(page, app, "e2e-conf")).toContainText("Mounted");

    await page.goto(appPage(app, "config-files"));
    await page
        .getByRole("row", { name: /E2E_CONF/ })
        .getByRole("button", { name: "Edit app config file" })
        .click();
    await page.getByRole("group", { name: "Value *" }).getByRole("textbox").fill(after);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.locator('[data-sonner-toast][data-type="success"]')).toBeVisible();
    await expectPrinted(page, app, "conf", after);

    await switchMount(page, app, "e2e-conf", "Disable");
    await expectPrinted(page, app, "conf", "missing");
    await expect(await mountRow(page, app, "e2e-conf")).toContainText("Not mounted");

    await switchMount(page, app, "e2e-conf", "Enable");
    await expectPrinted(page, app, "conf", after);

    expect((await latestDeployment(api, app))?.id, "no deployment was made").toBe(deployment?.id);
});

test("a certificate and its key mounted from an SSL certificate, the key 0400; renewed, the container has the new pair without a deploy", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "mount-tls");
    // A certificate is named by its domain.
    const domain = `${e2eName("tls")}.localhost`;
    await createSettingAt(api, `projects/${app.projectId}/ssl-certs`, {
        certType: "self-signed",
        domain,
        inheritable: true,
    });

    await newSettingMount(page, app, {
        name: "tls",
        from: "SSL certificate",
        setting: domain,
        files: [
            { part: "certificate", path: "/etc/tls/cert.pem" },
            { part: "privateKey", path: "/etc/tls/key.pem", mode: "0400" },
        ],
    });
    await expect(page.getByRole("row", { name: /tls/ })).toBeVisible();

    // Each file's first line, its letters only, and the start of its hash.
    const show = (key: string, file: string, more = "") =>
        `echo "${key}=$(head -1 ${file} | tr -dc A-Z)-$(sha256sum < ${file} | cut -c1-12)${more}"`;
    await deployImage(
        api,
        app,
        BUSYBOX,
        `sh -c 'while :; do ${show("cert", "/etc/tls/cert.pem")}; ` +
            `${show("key", "/etc/tls/key.pem", "-$(stat -c %a /etc/tls/key.pem)")}; sleep 2; done'`,
    );
    await deployed(api, app);
    const deployment = await latestDeployment(api, app);
    const cert = await expectPrinted(page, app, "cert", printed => /^BEGINCERTIFICATE-\w{12}$/.test(printed));
    const key = await expectPrinted(page, app, "key", printed => /^BEGIN\w*PRIVATEKEY-\w{12}-400$/.test(printed));

    await page.goto(`/projects/${app.projectId}/integrations/ssl-certificates/`);
    await page
        .getByRole("row", { name: new RegExp(domain) })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: "Renew Now" }).click();
    await expect(page.getByText("SSL certificate renewal started")).toBeVisible();

    await expectPrinted(page, app, "cert", printed => /^BEGINCERTIFICATE-\w{12}$/.test(printed) && printed !== cert);
    await expectPrinted(
        page,
        app,
        "key",
        printed => /^BEGIN\w*PRIVATEKEY-\w{12}-400$/.test(printed) && printed !== key,
    );
    expect((await latestDeployment(api, app))?.id, "no deployment was made").toBe(deployment?.id);
});

test("a basic auth's htpasswd mounted is one Apache's htpasswd accepts the password by, and no other; one the app cannot use is not offered", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "mount-htpasswd");
    const authName = e2eName("auth");
    const password = e2eName("password");
    await createSettingAt(api, `projects/${app.projectId}/${app.env}/basic-auth`, {
        name: authName,
        username: "e2e-user",
        password,
        inheritable: true,
    });
    // The env's own, not inheritable: its apps cannot use it.
    const envOnly = e2eName("env-only");
    await createSettingAt(api, `projects/${app.projectId}/${app.env}/basic-auth`, {
        name: envOnly,
        username: "e2e-user",
        password,
    });

    await page.goto(appPage(app, "setting-mounts"));
    await page.getByRole("button", { name: "New Setting Mount" }).click();
    await page.getByRole("combobox", { name: "Mount From" }).click();
    await page.getByRole("option", { name: "Basic auth", exact: true }).click();
    await page.getByRole("group", { name: "Setting *" }).getByRole("combobox").click();
    await expect(page.getByRole("option", { name: authName, exact: true })).toBeVisible();
    await expect(page.getByRole("option", { name: envOnly, exact: true })).toHaveCount(0);

    await newSettingMount(page, app, {
        name: "auth",
        from: "Basic auth",
        setting: authName,
        files: [{ part: "htpasswd", path: "/etc/app/htpasswd" }],
    });
    await expect(page.getByRole("row", { name: /auth/ })).toBeVisible();

    await deployImage(
        api,
        app,
        HTTPD,
        `sh -c 'htpasswd -vb /etc/app/htpasswd e2e-user ${password}; ` +
            `htpasswd -vb /etc/app/htpasswd e2e-user not-${password}; exec sleep 3600'`,
    );
    await deployed(api, app);
    await expectLogs(page, app, "Password for user e2e-user correct.");
    await expectLogs(page, app, "password verification failed");
});

test("a path one setting mount has is refused to another, and a config file a mount reads is not deleted", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "mount-refused");
    await createSettingAt(api, `${appPath(app)}/config-files`, { name: "E2E_CONF", content: e2eName("conf") });
    const file = { part: "content", path: "/etc/e2e.conf" };

    await newSettingMount(page, app, { name: "e2e-first", from: "Config file", setting: "E2E_CONF", files: [file] });
    await expect(page.getByRole("row", { name: /e2e-first/ })).toBeVisible();

    await newSettingMount(page, app, { name: "e2e-second", from: "Config file", setting: "E2E_CONF", files: [file] });
    await expect(page.getByText("'/etc/e2e.conf' is already used by the app's setting mount e2e-first")).toBeVisible();

    await page.goto(appPage(app, "config-files"));
    await page
        .getByRole("row", { name: /E2E_CONF/ })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: "Remove" }).click();
    await page.getByRole("dialog", { name: "Delete Item" }).getByRole("button", { name: "Delete" }).click();
    const inUse = page.getByRole("dialog", { name: "This setting is still in use" });
    await expect(inUse).toContainText("It cannot be deleted while anything still points at it.");
    // What reads it is listed, a link to the page it is changed on, which opens
    // in a tab of its own.
    const [mounts] = await Promise.all([
        page.context().waitForEvent("page"),
        inUse.getByRole("link", { name: /e2e-first/ }).click(),
    ]);
    await expect(mounts).toHaveURL(new RegExp(`/projects/${app.projectId}/[^/]+/apps/${app.id}/setting-mounts/$`));
    await expect(mounts.getByRole("row", { name: /e2e-first/ })).toBeVisible();
    await mounts.close();
    // Its footer's Close, or the corner's: either closes it.
    await inUse.getByRole("button", { name: "Close" }).first().click();

    await page.goto(appPage(app, "config-files"));
    await expect(page.getByRole("row", { name: /E2E_CONF/ })).toBeVisible();
});
