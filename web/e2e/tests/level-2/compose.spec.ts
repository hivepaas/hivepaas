import { type App, deleteProject, findAppNamed, findProject } from "../../support/api";
import { deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { visit } from "../../support/routing";

// The proxy's certificate for a .localhost name is its own, signed by no one.
test.use({ ignoreHTTPSErrors: true });

// Two services: one that serves at a published port, one that prints what its
// .env gave it. $$ is a dollar the container's shell reads, not one Compose
// fills in.
const COMPOSE = `services:
  web:
    image: traefik/whoami:latest
    ports: ["8080:80"]
  printer:
    image: busybox:1.37
    command: ["sh", "-c", "echo GREETING=$$GREETING; sleep 3600"]
    environment:
      GREETING: \${GREETING}
`;

test("a project made from a Compose file runs its services, with what its .env gives them, at a domain", async ({
    page,
    api,
    cleanup,
}) => {
    const name = e2eName("compose");
    cleanup(async () => {
        const made = await findProject(api, name);
        if (made) await deleteProject(api, made.id);
    });

    await page.goto("/projects/new/compose/");
    await page.getByRole("group", { name: "Compose File" }).getByRole("textbox").fill(COMPOSE);
    await page.getByRole("group", { name: ".env" }).getByRole("textbox").fill("GREETING=hello-from-compose\n");
    await page.getByRole("group", { name: "Project Name" }).getByRole("textbox").fill(name);
    // What the file becomes, shown before anything is made.
    await expect(page.getByRole("checkbox", { name: "Import printer" })).toBeChecked();
    await expect(page.getByRole("checkbox", { name: "Import web" })).toBeChecked();
    await expect(page.getByRole("checkbox", { name: "Deploy the apps once they are created" })).toBeChecked();
    // A published port becomes a domain: the one suggested, left as it is.
    const domain = await page.getByRole("group", { name: "Services" }).getByRole("textbox").getAttribute("placeholder");
    expect(domain).toMatch(/\.localhost$/);
    await page.getByRole("button", { name: "Create project" }).click();

    await expect(page.getByRole("link", { name: /^Open the/ })).toBeVisible({ timeout: 60_000 });
    const project = await findProject(api, name);
    expect(project, `project ${name} is made`).toBeDefined();
    const near: App = { id: "", name: "", projectId: project!.id, env: "prod" };
    const printer = await findAppNamed(api, near, "printer");
    const web = await findAppNamed(api, near, "web");
    expect(printer && web, "both services are apps").toBeTruthy();
    await deployed(api, web!);
    await deployed(api, printer!);
    await expectLogs(page, printer!, "GREETING=hello-from-compose");

    await visit(page, `https://${domain}/`);
    await expect(page.locator("body")).toContainText(`Host: ${domain}`);
});
