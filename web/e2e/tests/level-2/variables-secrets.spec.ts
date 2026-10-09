import type { APIRequestContext, Page } from "@playwright/test";

import {
    type App,
    type EnvVar,
    appPath,
    createSettingAt,
    deployImage,
    latestDeployment,
    setRuntimeEnvVars,
} from "../../support/api";
import { BUSYBOX, DEPLOYED, appIn, appPage, deployed, expectPrinted } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { REPOS, buildFromRepo } from "../../support/git";

test.describe.configure({ timeout: 300_000 });

const plain = (key: string, value: string): EnvVar => ({ key, value, isLiteral: false });

// printing prints, every two seconds, each variable as key=value: what the
// container was given is read in its log.
const printing = (...keys: string[]) =>
    `sh -c 'while :; do ${keys.map(key => `echo "${key.toLowerCase()}=$${key}"`).join("; ")}; sleep 2; done'`;

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
