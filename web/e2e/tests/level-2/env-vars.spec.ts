import { createApp, createProject, deleteProject, runtimeEnvVars } from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";

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
