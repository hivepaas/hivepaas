import type { APIRequestContext, Page } from "@playwright/test";
import crypto from "node:crypto";
import fs from "node:fs";

import {
    type App,
    type EnvVar,
    appPath,
    createApp,
    createSettingAt,
    deployImage,
    exposeApp,
    latestDeployment,
    setRuntimeEnvVars,
    settingIdNamed,
} from "../../support/api";
import { BUSYBOX, DEPLOYED, WHOAMI, appIn, appPage, deployed, expectPrinted } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { REPOS, buildFromRepo } from "../../support/git";

test.describe.configure({ timeout: 300_000 });

const plain = (key: string, value: string): EnvVar => ({ key, value, isLiteral: false });

// printing prints, every two seconds, each variable as key=value: what the
// container was given is read in its log.
const printing = (...keys: string[]) =>
    `sh -c 'while :; do ${keys.map(key => `echo "${key.toLowerCase()}=$${key}"`).join("; ")}; sleep 2; done'`;

// putSharedEnvVars replaces the variables an app shares with the apps of its
// env, its runtime ones left as they are.
async function putSharedEnvVars(api: APIRequestContext, app: App, vars: EnvVar[]): Promise<void> {
    const current = await api.get(`${appPath(app)}/env-vars`);
    const { updateVer, runtimeEnvVars } = (
        (await current.json()) as { data: { updateVer: number; runtimeEnvVars?: EnvVar[] | null } }
    ).data;
    const res = await api.put(`${appPath(app)}/env-vars`, {
        data: { updateVer, runtimeEnvVars: runtimeEnvVars ?? [], buildtimeEnvVars: [], sharedEnvVars: vars },
    });
    expect(res.ok(), `sharing ${app.name}'s variables: ${res.status()} ${await res.text()}`).toBe(true);
}

// variablesForm is the app's Runtime Env Variables, opened.
async function variablesForm(page: Page, app: App) {
    await page.goto(appPage(app, "env-variables"));
    return page.getByRole("region", { name: "Runtime Env Variables", exact: true });
}

// addVariable adds a runtime variable in the form, unsaved.
async function addVariable(page: Page, app: App, key: string, value: string): Promise<void> {
    const runtime = page.getByRole("region", { name: "Runtime Env Variables", exact: true });
    await runtime.getByRole("button", { name: "Add" }).click();
    await runtime.getByRole("group", { name: "New variable" }).getByRole("textbox", { name: "Key" }).fill(key);
    await runtime.getByRole("group", { name: key, exact: true }).getByRole("textbox", { name: "Value" }).fill(value);
}

// previewsOn turns the app's previews on in its Feature Settings.
async function previewsOn(api: APIRequestContext, app: App): Promise<void> {
    const current = await api.get(`${appPath(app)}/feature-settings`);
    const { updateVer } = ((await current.json()) as { data: { updateVer: number } }).data;
    const res = await api.put(`${appPath(app)}/feature-settings`, {
        data: { updateVer, previewSettings: { enabled: true, appsToClone: [], commands: [] } },
    });
    expect(res.ok(), `turning previews on: ${res.status()} ${await res.text()}`).toBe(true);
}

// previewOf makes a preview of the app's branch - an app of the same env, made
// by a task - and answers it once it is listed.
async function previewOf(api: APIRequestContext, app: App, branch: string): Promise<App> {
    const res = await api.post(`${appPath(app)}/previews`, { data: { repoRef: branch } });
    expect(res.ok(), `making a preview: ${res.status()} ${await res.text()}`).toBe(true);
    let id = "";
    await expect
        .poll(async () => {
            const list = await api.get(`${appPath(app)}/previews`);
            id = ((await list.json()) as { data: { id: string }[] }).data[0]?.id ?? "";
            return id;
        }, DEPLOYED)
        .not.toBe("");
    return { id, name: "preview", projectId: app.projectId, env: app.env };
}

test("a variable saved and a secret changed reach the running container without a deploy", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "applied");
    await createSettingAt(api, `${appPath(app)}/secrets`, { key: "E2E_TOKEN", value: "first-token" });
    await setRuntimeEnvVars(api, app, [plain("GREETING", "before"), plain("TOKEN", "${secrets.E2E_TOKEN}")]);
    await deployImage(api, app, BUSYBOX, printing("GREETING", "TOKEN"));
    await deployed(api, app);
    const deployment = await latestDeployment(api, app);
    await expectPrinted(page, app, "greeting", "before");
    await expectPrinted(page, app, "token", "first-token");

    const runtime = await variablesForm(page, app);
    await runtime
        .getByRole("group", { name: "GREETING", exact: true })
        .getByRole("textbox", { name: "Value" })
        .fill("after");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();
    await expectPrinted(page, app, "greeting", "after");

    await page.goto(appPage(app, "secrets"));
    await page
        .getByRole("row", { name: /E2E_TOKEN/ })
        .getByRole("button", { name: "Edit app secret" })
        .click();
    await page.getByPlaceholder("Leave empty to keep current value").fill("second-token");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.locator('[data-sonner-toast][data-type="success"]')).toBeVisible();
    await expectPrinted(page, app, "token", "second-token");

    expect((await latestDeployment(api, app))?.id, "no deployment was made").toBe(deployment?.id);
});

test("a preview takes its app's variables as they change, and goes without the secrets kept from previews", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "preview-vars");
    // Made through the API without saying: kept from the app's previews.
    await createSettingAt(api, `${appPath(app)}/secrets`, { key: "E2E_HELD", value: "held-token" });
    const vars = (greeting: string) => [plain("GREETING", greeting), plain("HELD", "${secrets.E2E_HELD}")];
    await setRuntimeEnvVars(api, app, vars("before"));
    await buildFromRepo(api, app, { repoURL: REPOS.shop, branch: "main", command: printing("GREETING", "HELD") });
    await deployed(api, app);
    await previewsOn(api, app);

    const preview = await previewOf(api, app, "main");
    await deployed(api, preview);
    await expectPrinted(page, preview, "greeting", "before");
    await expectPrinted(page, preview, "held", "");
    await expectPrinted(page, app, "held", "held-token");

    await setRuntimeEnvVars(api, app, vars("after"));
    await expectPrinted(page, app, "greeting", "after");
    await expectPrinted(page, preview, "greeting", "after");
});

test("variables that cannot be worked out are refused, and a secret a variable uses is not deleted", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "refused-vars");
    const save = async () => {
        await page.getByRole("button", { name: "Save" }).click();
    };

    await variablesForm(page, app);
    await addVariable(page, app, "E2E_MISSING", "${secrets.E2E_NOPE}");
    await save();
    await expect(page.getByText("secret 'E2E_NOPE' is missing")).toBeVisible();

    await variablesForm(page, app);
    await addVariable(page, app, "E2E_A", "${E2E_B}");
    await addVariable(page, app, "E2E_B", "${E2E_A}");
    await save();
    await expect(page.getByText(/Circular reference detected at 'E2E_[AB]'/)).toBeVisible();

    await createSettingAt(api, `${appPath(app)}/secrets`, { key: "E2E_KEPT", value: "kept" });
    await setRuntimeEnvVars(api, app, [plain("E2E_USES", "${secrets.E2E_KEPT}")]);
    await page.goto(appPage(app, "secrets"));
    await page
        .getByRole("row", { name: /E2E_KEPT/ })
        .getByRole("button", { name: "Actions menu" })
        .click();
    await page.getByRole("button", { name: "Remove" }).click();
    await page.getByRole("dialog", { name: "Delete Item" }).getByRole("button", { name: "Delete" }).click();
    await expect(page.getByText("secret 'E2E_KEPT' is missing")).toBeVisible();
    await page.goto(appPage(app, "secrets"));
    await expect(page.getByRole("row", { name: /E2E_KEPT/ })).toBeVisible();
});

test("a project's secret and an environment's config file made for its apps reach them; an environment's secret not made for them is refused", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "inherited");
    await createSettingAt(api, `projects/${app.projectId}/secrets`, {
        key: "E2E_SHARED",
        value: "shared-token",
        inheritable: true,
    });
    // The environment's own, made through the API without saying: not for its
    // apps.
    await createSettingAt(api, `projects/${app.projectId}/${app.env}/secrets`, {
        key: "E2E_HIDDEN",
        value: "hidden-token",
    });
    const conf = await createSettingAt(api, `projects/${app.projectId}/${app.env}/config-files`, {
        name: "E2E_ENV_CONF",
        content: "from-the-env",
        inheritable: true,
    });
    await createSettingAt(api, `${appPath(app)}/setting-mounts`, {
        name: "env-conf",
        source: { id: conf },
        files: [{ part: "content", path: "/etc/env.conf" }],
    });
    await setRuntimeEnvVars(api, app, [plain("SHARED", "${secrets.E2E_SHARED}")]);
    await deployImage(
        api,
        app,
        BUSYBOX,
        `sh -c 'while :; do echo "shared=$SHARED"; echo "conf=$(cat /etc/env.conf)"; sleep 2; done'`,
    );
    await deployed(api, app);
    await expectPrinted(page, app, "shared", "shared-token");
    await expectPrinted(page, app, "conf", "from-the-env");

    await variablesForm(page, app);
    await addVariable(page, app, "HIDDEN", "${secrets.E2E_HIDDEN}");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText(/secret 'E2E_HIDDEN' is not available here/)).toBeVisible();
});

test("a secret's value is not shown, and neither revealed nor downloaded while the server returns no secrets", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "masked");
    const value = e2eName("never-shown");

    await page.goto(appPage(app, "secrets"));
    await page.getByRole("button", { name: "New Secret" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("E2E_MASKED");
    await page.getByRole("group", { name: "Value *" }).getByRole("textbox").fill(value);
    await page.getByRole("button", { name: "Save" }).click();
    const row = page.getByRole("row", { name: /E2E_MASKED/ });
    await expect(row).toBeVisible();
    await expect(page.getByText(value)).toHaveCount(0);

    await row.getByRole("button", { name: "Edit app secret" }).click();
    await expect(page.getByPlaceholder("Leave empty to keep current value")).toHaveValue("");
    await expect(page.getByText(value)).toHaveCount(0);

    // The env's server keeps secrets to itself (see 2.26): a reveal is refused,
    // and a download, which hands the secret over as a reveal does, as well.
    await page.getByRole("button", { name: "Reveal Secret" }).click();
    await page
        .getByRole("dialog", { name: "Reveal Secret" })
        .getByRole("button", { name: "Reveal the secret" })
        .click();
    await expect(page.getByText("Returning secrets via the API is disabled on this server.")).toBeVisible();
    await expect(page.getByText(value)).toHaveCount(0);

    await page.goto(appPage(app, "secrets"));
    let opened = false;
    page.context().on("page", () => {
        opened = true;
    });
    await row.getByRole("button", { name: "Actions menu" }).click();
    await page.getByRole("button", { name: "Download File" }).click();
    await expect(page.getByText("Returning secrets via the API is disabled on this server.")).toBeVisible();
    expect(opened, "no tab opened with the secret").toBe(false);
});

// finalValues are the keys and values of the Final Env Values dialog.
async function finalValues(page: Page): Promise<Record<string, string>> {
    const dialog = page.getByRole("dialog", { name: "Final Env Values" });
    await expect(dialog.locator("textarea").first()).toBeVisible();
    return dialog.evaluate(element => {
        const values: Record<string, string> = {};
        for (const textarea of element.querySelectorAll("textarea")) {
            const key = textarea.parentElement?.querySelector("input");
            if (key) values[key.value] = textarea.value;
        }
        return values;
    });
}

test("a literal variable reaches the container as written and a reference worked out; the final values show both, a secret masked", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "literal");
    await createSettingAt(api, `${appPath(app)}/secrets`, { key: "E2E_TOKEN", value: "kept-secret" });
    await setRuntimeEnvVars(api, app, [
        plain("GREETING", "hello"),
        plain("REF", "${GREETING}"),
        { key: "LIT", value: "${GREETING}", isLiteral: true },
        plain("TOKEN", "${secrets.E2E_TOKEN}"),
    ]);
    await deployImage(api, app, BUSYBOX, printing("REF", "LIT"));
    await deployed(api, app);
    await expectPrinted(page, app, "ref", "hello");
    await expectPrinted(page, app, "lit", "${GREETING}");

    const runtime = await variablesForm(page, app);
    await runtime.getByRole("button", { name: "Show Final Values" }).click();
    expect(await finalValues(page)).toEqual(
        expect.objectContaining({ REF: "hello", LIT: "${GREETING}", TOKEN: "********" }),
    );
});

test("Link App: an app reaches another at the address it adds; a shared variable changed reaches it without a deploy", async ({
    page,
    api,
    cleanup,
}) => {
    const client = await appIn(api, cleanup, "linked");
    const target = await createApp(api, { id: client.projectId }, "api");
    await exposeApp(api, target, `${e2eName("api")}.localhost`);
    await putSharedEnvVars(api, target, [plain("API_TOKEN", "first")]);
    await deployImage(api, target, WHOAMI);
    await deployed(api, target);

    const runtime = await variablesForm(page, client);
    await runtime.getByRole("button", { name: "Link App" }).click();
    const dialog = page.getByRole("dialog", { name: "Link to another app" });
    await dialog.getByRole("combobox").first().click();
    await page.getByRole("option", { name: /^api/ }).click();
    // Its addresses are picked, as recommended; its shared variables too.
    await dialog
        .locator("div.rounded-md", { has: page.getByText("Shared variables of api", { exact: true }) })
        .getByRole("checkbox")
        .first()
        .check();
    await dialog.getByRole("button", { name: /^Add \d+ variables$/ }).click();
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();

    await deployImage(
        api,
        client,
        BUSYBOX,
        `sh -c 'while :; do echo "reached=$(wget -qO- -T 2 "$API_URL" 2>/dev/null | grep -c Hostname)"; ` +
            `echo "token=$API_TOKEN"; sleep 2; done'`,
    );
    await deployed(api, client);
    const deployment = await latestDeployment(api, client);
    await expectPrinted(page, client, "reached", "1");
    await expectPrinted(page, client, "token", "first");

    await putSharedEnvVars(api, target, [plain("API_TOKEN", "second")]);
    await expectPrinted(page, client, "token", "second");
    expect((await latestDeployment(api, client))?.id, "no deployment was made").toBe(deployment?.id);
});

test("a binary secret and a binary config file, uploaded, are mounted as they were; a config file downloads as it was", async ({
    page,
    api,
    cleanup,
}, testInfo) => {
    const app = await appIn(api, cleanup, "binary");
    const blob = Buffer.from(Array.from({ length: 1024 }, (_, i) => (i * 7) % 256));
    const file = testInfo.outputPath("blob.bin");
    fs.writeFileSync(file, blob);
    const sha = crypto.createHash("sha256").update(blob).digest("hex").slice(0, 16);

    for (const { tab, newButton, input, name } of [
        { tab: "secrets", newButton: "New Secret", input: "#app-secret-binary-value", name: "E2E_BLOB" },
        {
            tab: "config-files",
            newButton: "New Config File",
            input: "#app-config-file-binary-value",
            name: "E2E_BLOB_CONF",
        },
    ]) {
        await page.goto(appPage(app, tab));
        await page.getByRole("button", { name: newButton }).click();
        await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill(name);
        await page.getByRole("tab", { name: "Binary" }).click();
        await page.locator(input).setInputFiles(file);
        await page.getByRole("button", { name: "Save" }).click();
        await expect(page.getByRole("row", { name: new RegExp(name) })).toBeVisible();
    }
    const secret = await settingIdNamed(api, `${appPath(app)}/secrets`, "E2E_BLOB");
    const conf = await settingIdNamed(api, `${appPath(app)}/config-files`, "E2E_BLOB_CONF");
    await createSettingAt(api, `${appPath(app)}/setting-mounts`, {
        name: "blob-secret",
        source: { id: secret },
        files: [{ part: "value", path: "/etc/blob.secret" }],
    });
    await createSettingAt(api, `${appPath(app)}/setting-mounts`, {
        name: "blob-conf",
        source: { id: conf },
        files: [{ part: "content", path: "/etc/blob.conf" }],
    });
    await deployImage(
        api,
        app,
        BUSYBOX,
        `sh -c 'while :; do echo "secret=$(sha256sum < /etc/blob.secret | cut -c1-16)"; ` +
            `echo "conf=$(sha256sum < /etc/blob.conf | cut -c1-16)"; sleep 2; done'`,
    );
    await deployed(api, app);
    await expectPrinted(page, app, "secret", sha);
    await expectPrinted(page, app, "conf", sha);

    const text = e2eName("as-it-was");
    await createSettingAt(api, `${appPath(app)}/config-files`, { name: "E2E_TEXT", content: text });
    await page.goto(appPage(app, "config-files"));
    await page
        .getByRole("row", { name: /E2E_TEXT/ })
        .getByRole("button", { name: "Actions menu" })
        .click();
    const [download] = await Promise.all([
        page.context().waitForEvent("page"),
        page.getByRole("button", { name: "Download File" }).click(),
    ]);
    await expect(download.locator("body")).toHaveText(text);
    await download.close();
});
