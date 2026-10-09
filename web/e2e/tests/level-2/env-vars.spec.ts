import type { APIRequestContext } from "@playwright/test";

import {
    type App,
    type EnvVar,
    appPath,
    createApp,
    createProject,
    deleteProject,
    deployImage,
    runtimeEnvVars,
} from "../../support/api";
import { BUSYBOX, appPage, deployed, expectShownLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 240_000 });

const plain = (key: string, value: string): EnvVar => ({ key, value, isLiteral: false });

// putEnvVars replaces the runtime variables saved at a path - a project's, an
// environment's, an app's - and, for an app, those it shares.
async function putEnvVars(
    api: APIRequestContext,
    path: string,
    runtimeEnvVars: EnvVar[],
    sharedEnvVars?: EnvVar[],
): Promise<void> {
    const current = await api.get(`${path}/env-vars`);
    expect(current.ok(), `reading ${path}/env-vars: ${current.status()}`).toBe(true);
    // None saved yet: there is nothing to have changed since.
    const updateVer = ((await current.json()) as { data: { updateVer: number } | null }).data?.updateVer ?? 0;
    const res = await api.put(`${path}/env-vars`, {
        data: {
            updateVer,
            runtimeEnvVars,
            buildtimeEnvVars: [],
            ...(sharedEnvVars ? { sharedEnvVars } : {}),
        },
    });
    expect(res.ok(), `saving ${path}/env-vars: ${res.status()} ${await res.text()}`).toBe(true);
}

// Each row of the variables' form is named by its key, and each field and
// button in it by what it is: a variable is found, filled and removed by name.
test("an env variable is added, kept after a reload, and removed", async ({ page, api, cleanup }) => {
    const project = await createProject(api, e2eName("env-vars"));
    cleanup(() => deleteProject(api, project.id));
    const app = await createApp(api, project, "web");

    await page.goto(`/projects/${app.projectId}/${app.env}/apps/${app.id}/env-variables/`);
    const runtime = page.getByRole("region", { name: "Runtime Env Variables", exact: true });
    await runtime.getByRole("button", { name: "Add" }).click();
    await runtime
        .getByRole("group", { name: "New variable" })
        .getByRole("textbox", { name: "Key" })
        .fill("E2E_PATTERN");

    // Literal: its reference is kept as written, not replaced.
    const row = runtime.getByRole("group", { name: "E2E_PATTERN", exact: true });
    await row.getByRole("textbox", { name: "Value" }).fill("${HOME}");
    await row.getByRole("checkbox", { name: "Literal" }).check();
    const multiline = row.getByRole("button", { name: "Multi-line value" });
    await multiline.click();
    await expect(multiline).toHaveAttribute("aria-pressed", "true");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();
    expect(await runtimeEnvVars(api, app)).toEqual([
        expect.objectContaining({ key: "E2E_PATTERN", value: "${HOME}", isLiteral: true }),
    ]);

    await page.reload();
    await expect(row.getByRole("textbox", { name: "Value" })).toHaveValue("${HOME}");
    await expect(row.getByRole("checkbox", { name: "Literal" })).toBeChecked();

    await row.getByRole("button", { name: "Remove variable" }).click();
    await expect(row).toHaveCount(0);
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Environment variables updated")).toBeVisible();
    expect(await runtimeEnvVars(api, app)).toEqual([]);
});

// An app's container gets the variables of its project and its environment -
// the environment's over the project's, its own over both - another app's it
// refers to, and those HivePaaS sets. Its page lists the inherited ones.
test("variables of the project, the environment and another app reach the container, the nearest winning", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("env-inherit"));
    cleanup(() => deleteProject(api, project.id));
    await putEnvVars(api, `projects/${project.id}`, [
        plain("PROJECT_ONLY", "from-project"),
        plain("BOTH", "from-project"),
        plain("APP_WINS", "from-project"),
    ]);
    await putEnvVars(api, `projects/${project.id}/development`, [
        plain("ENV_ONLY", "from-env"),
        plain("BOTH", "from-env"),
    ]);
    const db = await createApp(api, project, "db");
    await putEnvVars(api, appPath(db), [], [plain("DB_NAME", "shop")]);
    const web: App = await createApp(api, project, "web");
    await putEnvVars(api, appPath(web), [
        plain("APP_WINS", "from-app"),
        plain("LINKED", "${db.DB_NAME}"),
        plain("NAMED", "${HIVEPAAS_APP_NAME}-in-${HIVEPAAS_ENV}"),
    ]);

    await deployImage(api, web, BUSYBOX, "sh -c 'env | sort; exec sleep 3600'");
    await deployed(api, web);
    await page.goto(appPage(web, "logs"));
    for (const line of [
        "PROJECT_ONLY=from-project",
        "ENV_ONLY=from-env",
        "BOTH=from-env",
        "APP_WINS=from-app",
        "LINKED=shop",
        "NAMED=web-in-development",
        "HIVEPAAS_HOST=web",
    ]) {
        await expectShownLogs(page, line);
    }

    await page.goto(appPage(web, "env-variables"));
    const inherited = page.getByRole("region", { name: "Inherited Runtime Env Variables" });
    for (const [key, value] of [
        ["PROJECT_ONLY", "from-project"],
        ["ENV_ONLY", "from-env"],
        ["BOTH", "from-env"],
    ] as const) {
        await expect(
            inherited.getByRole("group", { name: key, exact: true }).getByRole("textbox", { name: "Value" }),
        ).toHaveValue(value);
    }
});
