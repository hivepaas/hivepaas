import { createProject, deleteProject } from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";

test("making and deleting a project is in the audit log, with who did it", async ({ page, api }) => {
    const me = (await (await api.get("sessions/me")).json()) as { data: { user: { username: string } } };
    const project = await createProject(api, e2eName("audited"));
    await deleteProject(api, project.id);
    await page.goto("/operations/audit-logs/");

    await page.getByRole("searchbox", { name: "Search" }).fill(project.name);

    // The search is applied: what is listed is the project's alone, not the
    // newest entries of anything, among which the project's may happen to be.
    const listed = page.getByRole("button").filter({ hasText: /^(Allowed|Denied)/ });
    await expect(listed.filter({ hasNotText: project.name })).toHaveCount(0);
    const entries = listed.filter({ hasText: project.name });
    for (const action of ["project-create", "project-delete"]) {
        const entry = entries.filter({ hasText: action });
        await expect(entry).toHaveCount(1);
        await expect(entry).toContainText(me.data.user.username);
    }
});
