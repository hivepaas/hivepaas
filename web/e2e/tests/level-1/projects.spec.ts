import { createProject, deleteProject, findProject } from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";

test.describe("projects", () => {
    test("a project made from the dialog is listed, with the two default environments", async ({
        page,
        api,
        cleanup,
    }) => {
        const name = e2eName("project-new");
        cleanup(async () => {
            const made = await findProject(api, name);
            if (made) await deleteProject(api, made.id);
        });
        await page.goto("/projects/");

        await page.getByRole("button", { name: "New Project", exact: true }).click();
        await page.getByRole("menuitem", { name: "Empty project" }).click();
        const dialog = page.getByRole("dialog", { name: "Create Project" });
        await dialog.getByRole("textbox", { name: "Name *" }).fill(name);
        await dialog.getByRole("button", { name: "Create Project" }).click();

        await expect(page.getByText("Project created successfully")).toBeVisible();
        await expect(page.getByRole("row").filter({ hasText: name })).toBeVisible();
        const made = await findProject(api, name);
        expect(made?.envs?.map(env => env.name)).toEqual(["development", "production"]);
    });

    test("a project needs a name", async ({ page }) => {
        await page.goto("/projects/");

        await page.getByRole("button", { name: "New Project", exact: true }).click();
        await page.getByRole("menuitem", { name: "Empty project" }).click();
        const dialog = page.getByRole("dialog", { name: "Create Project" });
        await dialog.getByRole("button", { name: "Create Project" }).click();

        await expect(dialog.getByRole("alert")).toHaveText("Name is required");
        await expect(dialog).toBeVisible();
    });

    test("a second project of the same name is refused", async ({ page, api, cleanup }) => {
        const existing = await createProject(api, e2eName("project-taken"));
        cleanup(() => deleteProject(api, existing.id));
        await page.goto("/projects/");

        await page.getByRole("button", { name: "New Project", exact: true }).click();
        await page.getByRole("menuitem", { name: "Empty project" }).click();
        const dialog = page.getByRole("dialog", { name: "Create Project" });
        await dialog.getByRole("textbox", { name: "Name *" }).fill(existing.name);
        await dialog.getByRole("button", { name: "Create Project" }).click();

        await expect(page.getByText("Project already exists")).toBeVisible();
        await expect(dialog).toBeVisible();
    });

    test("renaming a project and adding an environment are kept", async ({ page, api, cleanup }) => {
        const project = await createProject(api, e2eName("project-edit"));
        cleanup(() => deleteProject(api, project.id));
        const renamed = `${project.name}-renamed`;
        await page.goto(`/projects/${project.id}/settings/general/`);

        await page.getByRole("textbox", { name: "Enter project name" }).fill(renamed);
        await page.getByRole("button", { name: "Add environment" }).click();
        await page.getByRole("textbox", { name: "Enter env" }).fill("staging");
        await page.getByRole("textbox", { name: "Enter env" }).press("Enter");
        await page.getByRole("button", { name: "Save" }).click();

        await expect(page.getByText("Project information updated")).toBeVisible();
        await page.reload();
        await expect(page.getByRole("heading", { name: renamed })).toBeVisible();
        await expect(page.getByRole("button", { name: "Remove staging" })).toBeVisible();
        const saved = await findProject(api, renamed);
        expect(saved?.envs?.map(env => env.name)).toEqual(["development", "production", "staging"]);
    });

    test("deleting a project asks for its name, then removes it", async ({ page, api, cleanup }) => {
        const project = await createProject(api, e2eName("project-gone"));
        cleanup(() => deleteProject(api, project.id));
        await page.goto(`/projects/${project.id}/settings/general/`);

        await page.getByRole("tab", { name: "Danger Zone" }).click();
        await page.getByRole("button", { name: "Delete Project" }).click();
        const dialog = page.getByRole("dialog", { name: "Delete project" });
        const confirm = dialog.getByRole("button", { name: "Delete this Project" });
        await expect(confirm).toBeDisabled();
        await dialog.getByRole("textbox").fill(project.name);
        // Its data too: a run leaves nothing behind.
        await dialog.getByRole("checkbox", { name: /Also delete the stored data/ }).check();
        await confirm.click();

        await expect(page).toHaveURL(/\/projects\/$/);
        await expect.poll(() => findProject(api, project.name)).toBeUndefined();
    });
});
