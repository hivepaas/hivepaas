import type { APIRequestContext } from "@playwright/test";

import {
    appPath,
    createApp,
    createProject,
    createSettingAt,
    deleteProject,
    deployImage,
    findAppNamed,
    findProject,
    setRuntimeEnvVars,
} from "../../../support/api";
import { BUSYBOX, deployed, expectPrinted } from "../../../support/apps";
import { env } from "../../../support/env";
import { e2eName, expect, test } from "../../../support/fixtures";

// These run alone, after every other test (playwright.config.ts): the server
// returns secrets through its API while they run.
test.describe.configure({ mode: "serial", timeout: 300_000 });

interface SecuritySettings {
    returnSecretsViaApi: boolean;
    alwaysReturnSecretTypes: string[] | null;
    allowPrivilegedApps: boolean;
}

// returnSecrets switches the server's return of secrets through its API, as
// System's security settings do: the app secret re-entered, the other switches
// as they are.
async function returnSecrets(api: APIRequestContext, on: boolean): Promise<void> {
    const current = await api.get("system/hivepaas/security-settings");
    expect(current.ok(), `reading the security settings: ${current.status()}`).toBe(true);
    const settings = ((await current.json()) as { data: SecuritySettings }).data;
    const res = await api.put("system/hivepaas/security-settings", {
        data: {
            appSecret: env.appSecret,
            returnSecretsViaApi: on,
            alwaysReturnSecretTypes: settings.alwaysReturnSecretTypes ?? [],
            allowPrivilegedApps: settings.allowPrivilegedApps,
        },
    });
    expect(res.ok(), `turning the return of secrets ${on ? "on" : "off"}: ${res.status()} ${await res.text()}`).toBe(
        true,
    );
}

test.beforeAll(async ({ api }) => {
    await returnSecrets(api, true);
});

test.afterAll(async ({ api }) => {
    await returnSecrets(api, false);
});

// A project's secrets travel in its bundle encrypted with a passphrase: the
// bundle imported is opened with it - another is refused - and the app it
// restores runs with its secret as it was.
test("a project exported with its secrets encrypted comes back with them, opened by the passphrase and no other", async ({
    page,
    api,
    cleanup,
}) => {
    const name = e2eName("spec-secrets");
    cleanup(async () => {
        const made = await findProject(api, name);
        if (made) await deleteProject(api, made.id);
    });
    const token = e2eName("token");
    const passphrase = e2eName("bundle-passphrase");
    const project = await createProject(api, name);
    const app = await createApp(api, project, "web");
    await createSettingAt(api, `${appPath(app)}/secrets`, { key: "E2E_TOKEN", value: token });
    await setRuntimeEnvVars(api, app, [{ key: "TOKEN", value: "${secrets.E2E_TOKEN}", isLiteral: false }]);
    await deployImage(api, app, BUSYBOX, `sh -c 'while :; do echo "token=$TOKEN"; sleep 2; done'`);
    await deployed(api, app);

    await page.goto(`/projects/${project.id}/operations/export/`);
    await page.getByRole("group", { name: "Secrets" }).getByRole("combobox").click();
    await page.getByRole("option", { name: "Include, encrypted" }).click();
    await page.getByRole("group", { name: "Passphrase" }).locator("input").fill(passphrase);
    const downloading = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export", exact: true }).click();
    const bundle = test.info().outputPath("bundle.tar.gz");
    await (await downloading).saveAs(bundle);
    await deleteProject(api, project.id);

    await page.goto("/operations/export/");
    const choosing = page.waitForEvent("filechooser");
    await page.getByRole("button", { name: "Choose file" }).click();
    await (await choosing).setFiles(bundle);
    const passphraseField = page.getByRole("group", { name: "Passphrase" }).locator("input");
    await passphraseField.fill(`not-${passphrase}`);
    await page.getByRole("button", { name: "Open", exact: true }).click();
    // The bundle is read again with each passphrase: a while.
    await expect(page.getByText("The passphrase does not open this bundle")).toBeVisible({ timeout: 60_000 });
    await passphraseField.fill(passphrase);
    await page.getByRole("button", { name: "Open", exact: true }).click();
    await expect(page.getByRole("checkbox", { name: "Import web" })).toBeChecked({ timeout: 60_000 });
    await page.getByRole("button", { name: /^Import( and accept \d+ issues?)?$/ }).click();
    await expect(page.getByText("Imported", { exact: true })).toBeVisible({ timeout: 60_000 });

    const restored = await findProject(api, name);
    expect(restored?.id, "a project of its own again").not.toBe(project.id);
    const web = await findAppNamed(api, { id: "", name: "", projectId: restored?.id ?? "", env: "development" }, "web");
    expect(web, "its app").toBeDefined();
    if (!web) return;
    await deployed(api, web);
    await expectPrinted(page, web, "token", token);
});
