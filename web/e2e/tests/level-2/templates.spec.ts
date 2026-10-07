import { type App, createProject, deleteProject, findAppNamed } from "../../support/api";
import { deployed, expectInstances } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

// The catalog is the app-templates checkout env/up.sh gives the backend.
test("an app made from a template is deployed, and runs", async ({ page, api, cleanup }) => {
    const project = await createProject(api, e2eName("template"));
    cleanup(() => deleteProject(api, project.id));

    await page.goto(`/projects/${project.id}/app-templates/`);
    await page.getByRole("textbox", { name: /^Search templates/ }).fill("IT Tools");
    await page
        .getByRole("button", { name: /^IT Tools/ })
        .getByRole("button", { name: "Deploy", exact: true })
        .click();
    const dialog = page.getByRole("dialog", { name: "Deploy Template: IT Tools" });
    await expect(dialog.getByRole("textbox", { name: "e.g. my-app" })).toHaveValue("it-tools");
    await dialog.getByRole("button", { name: "Deploy Template" }).click();

    const near: App = { id: "", name: "", projectId: project.id, env: "development" };
    await expect.poll(() => findAppNamed(api, near, "it-tools"), { timeout: 60_000 }).toBeDefined();
    const app = (await findAppNamed(api, near, "it-tools"))!;
    await deployed(api, app);
    // Its health check passes: the page is served.
    await expectInstances(page, app, "1/1");
});
