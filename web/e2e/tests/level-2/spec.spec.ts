import {
    createApp,
    createProject,
    deleteProject,
    deployImage,
    exposeApp,
    findAppNamed,
    findProject,
    runtimeEnvVars,
    setRuntimeEnvVars,
} from "../../support/api";
import { WHOAMI, deployed } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { domainFor, visit } from "../../support/routing";

// Each test deploys, exports, imports and deploys again: room for it.
test.describe.configure({ timeout: 180_000 });
// The proxy's certificate for a .localhost name is its own, signed by no one.
test.use({ ignoreHTTPSErrors: true });

// Restoring a project from its spec: exported from its Operations page,
// deleted, and imported on the installation's. What it is made of comes back,
// and runs.
test("a project exported, deleted and imported again runs as it was: its app, its variable, its domain", async ({
    page,
    api,
    cleanup,
}) => {
    const name = e2eName("spec");
    cleanup(async () => {
        const made = await findProject(api, name);
        if (made) await deleteProject(api, made.id);
    });
    const project = await createProject(api, name);
    const app = await createApp(api, project, "web");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    await setRuntimeEnvVars(api, app, [{ key: "E2E_SPEC", value: "kept", isLiteral: false }]);
    const domain = domainFor("spec");
    await exposeApp(api, app, domain);

    await page.goto(`/projects/${project.id}/operations/export/`);
    const downloading = page.waitForEvent("download");
    await page.getByRole("button", { name: "Export", exact: true }).click();
    const bundle = test.info().outputPath("bundle.tar.gz");
    await (await downloading).saveAs(bundle);

    await deleteProject(api, project.id);

    await page.goto("/operations/export/");
    const choosing = page.waitForEvent("filechooser");
    await page.getByRole("button", { name: "Choose file" }).click();
    await (await choosing).setFiles(bundle);
    await expect(page.getByRole("checkbox", { name: new RegExp(`^Import ${name} `) })).toBeChecked();
    await expect(page.getByRole("checkbox", { name: "Import web" })).toBeChecked();
    await expect(page.getByRole("checkbox", { name: "Deploy the apps it creates" })).toBeChecked();
    await page.getByRole("button", { name: /^Import( and accept \d+ issues?)?$/ }).click();
    await expect(page.getByText("Imported", { exact: true })).toBeVisible({ timeout: 60_000 });

    const restored = await findProject(api, name);
    expect(restored?.id, "a project of its own again").not.toBe(project.id);
    const web = await findAppNamed(api, { id: "", name: "", projectId: restored!.id, env: "development" }, "web");
    expect(web, "its app").toBeDefined();
    await deployed(api, web!);
    expect(await runtimeEnvVars(api, web!)).toContainEqual(expect.objectContaining({ key: "E2E_SPEC", value: "kept" }));
    await visit(page, `https://${domain}/`);
    await expect(page.locator("body")).toContainText(`Host: ${domain}`);
});

// A secret travels only in a bundle the server lets out: by default it does
// not return secrets through the API. Asking for them is refused, and the page
// says why and where that is changed.
test("exporting secrets is refused while the server does not return them, and says so", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("spec-secret"));
    cleanup(() => deleteProject(api, project.id));

    await page.goto(`/projects/${project.id}/operations/export/`);
    await page.getByRole("group", { name: "Secrets" }).getByRole("combobox").click();
    await page.getByRole("option", { name: "Include, encrypted" }).click();
    await page.getByRole("group", { name: "Passphrase" }).locator("input").fill("e2e-bundle-passphrase");
    await page.getByRole("button", { name: "Export", exact: true }).click();

    await expect(
        page.getByText(
            "Returning secrets via the API is disabled on this server. Enable it in the System configuration.",
        ),
    ).toBeVisible();
});
