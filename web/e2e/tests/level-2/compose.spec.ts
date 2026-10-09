import type { APIRequestContext } from "@playwright/test";
import fs from "node:fs";
import path from "node:path";

import {
    type App,
    createApp,
    createProject,
    deleteProject,
    deployImage,
    findAppNamed,
    findProject,
    latestDeployment,
} from "../../support/api";
import { WHOAMI, deployed, expectLogs } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { domainFor, visit } from "../../support/routing";

test.describe.configure({ timeout: 240_000 });

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

// appKeyed is the app of the environment known by that key, if there is one:
// an app created under another key keeps its service's name.
async function appKeyed(api: APIRequestContext, near: App, key: string): Promise<App | undefined> {
    const res = await api.get(`projects/${near.projectId}/${near.env}/apps`);
    expect(res.ok(), `listing apps: ${res.status()}`).toBe(true);
    const found = ((await res.json()) as { data: { id: string; name: string; key: string }[] }).data.find(
        a => a.key === key,
    );
    return found ? { id: found.id, name: found.name, projectId: near.projectId, env: near.env } : undefined;
}

// The env has apps of both services' names: web, which serves, and worker. The
// file's web is the env's own; its worker is created under another key, and
// reaches web by name - the env's app.
const INTO_EXISTING = `services:
  web:
    image: traefik/whoami:latest
  worker:
    image: busybox:1.37
    command: ["sh", "-c", "while true; do wget -qO- http://web/ | head -n 1; sleep 3; done"]
`;

test("a Compose file added to a project uses the env's app of a service's name, or creates it under another key", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("compose-into"));
    cleanup(() => deleteProject(api, project.id));
    const web = await createApp(api, project, "web");
    await deployImage(api, web, WHOAMI);
    await deployed(api, web);
    const worker = await createApp(api, project, "worker");
    const webDeployment = (await latestDeployment(api, web))?.id;

    await page.goto(`/projects/${project.id}/apps/from-compose/`);
    await expect(page.getByRole("group", { name: "Environment" }).getByRole("combobox")).toHaveText("development");
    await page.getByRole("group", { name: "Compose File" }).getByRole("textbox").fill(INTO_EXISTING);
    // Nothing is added while a service's name is an app's of the env.
    await expect(page.getByText("2 issues block adding the apps.")).toBeVisible();
    await page.getByRole("combobox", { name: "What to do with web" }).click();
    await page.getByRole("option", { name: "Use the env's app, as it is" }).click();
    await page.getByRole("combobox", { name: "What to do with worker" }).click();
    await page.getByRole("option", { name: "Create it under another key" }).click();
    await expect(page.getByRole("textbox", { name: "The key worker is created under" })).toHaveValue("worker-2");
    await page.getByRole("button", { name: "Add apps" }).click();

    // Added cleanly: the project's apps, in the env they went into.
    await expect(page).toHaveURL(new RegExp(`/projects/${project.id}/apps/$`), { timeout: 60_000 });
    const added = await appKeyed(api, web, "worker-2");
    expect(added, "worker is created as worker-2").toBeDefined();
    expect(added!.name, "named as its service").toBe("worker");
    await deployed(api, added!);
    // web answers worker-2: by its name, the env's web - whose container's
    // name is its key.
    await expectLogs(page, added!, "Hostname: web");

    // What the env had is as it was: web not deployed again, worker still
    // there, no web created beside them.
    expect((await latestDeployment(api, web))?.id, "web is not deployed again").toBe(webDeployment);
    expect(await appKeyed(api, web, "worker"), "the env's worker stays").toEqual(
        expect.objectContaining({ id: worker.id }),
    );
    expect(await appKeyed(api, web, "web-2"), "no second web").toBeUndefined();
    expect(await findAppNamed(api, web, "web"), "web is the env's own").toEqual(
        expect.objectContaining({ id: web.id }),
    );
});

// A compose project's folder, as it is checked out: the compose file, and what
// it reads beside it - a secret's file, a file it mounts, a directory of files
// it serves. Its site is routed by Traefik labels, as on a server of its own.
function composeFolder(dir: string, domain: string): void {
    const files: Record<string, string> = {
        "compose.yaml": `services:
  site:
    image: busybox:1.37
    command: ["busybox", "httpd", "-f", "-p", "80", "-h", "/www"]
    volumes:
      - ./www:/www:ro
    labels:
      - traefik.enable=true
      - traefik.http.routers.site.rule=Host(\`${domain}\`)
      - traefik.http.services.site.loadbalancer.server.port=80
  printer:
    image: busybox:1.37
    command: ["sh", "-c", "echo TOKEN=$$(cat /run/secrets/token); echo CONF=$$(cat /etc/app.conf); sleep 3600"]
    volumes:
      - ./app.conf:/etc/app.conf:ro
    secrets:
      - token
secrets:
  token:
    file: ./token.txt
`,
        "token.txt": "token-from-the-folder",
        "app.conf": "conf-from-the-folder",
        "www/index.html": "<h1>site-from-the-folder</h1>",
    };
    for (const [name, content] of Object.entries(files)) {
        fs.mkdirSync(path.dirname(path.join(dir, name)), { recursive: true });
        fs.writeFileSync(path.join(dir, name), content);
    }
}

test("a Compose folder opened gives its services the files they read, and its Traefik labels a domain", async ({
    page,
    api,
    cleanup,
}) => {
    const name = e2eName("compose-folder");
    cleanup(async () => {
        const made = await findProject(api, name);
        if (made) await deleteProject(api, made.id);
    });
    const domain = domainFor("compose-labels");
    const dir = test.info().outputPath("site");
    composeFolder(dir, domain);

    await page.goto("/projects/new/compose/");
    const chooser = page.waitForEvent("filechooser");
    await page.getByRole("button", { name: "Open folder" }).click();
    await (await chooser).setFiles(dir);
    await expect(page.getByText(/^site\/compose\.yaml and the 3 files beside it/)).toBeVisible();
    await page.getByRole("group", { name: "Project Name" }).getByRole("textbox").fill(name);

    // What the file reads is taken from the folder: nothing is missing.
    const needs = page.getByRole("group", { name: "Files" });
    await expect(needs).toContainText("token.txt");
    await expect(needs).toContainText("app.conf");
    await expect(needs.getByRole("checkbox", { name: "Mount the folder's 1 file in it, read only" })).toBeChecked();
    await expect(needs).not.toContainText("missing");
    // The labels' host is the site's domain.
    await expect(page.getByText("From the service's Traefik labels.")).toBeVisible();
    await expect(page.getByRole("group", { name: "Services" }).getByRole("textbox")).toHaveValue(domain);
    await page.getByRole("button", { name: /^Create/ }).click();

    await expect(page.getByRole("link", { name: /^Open the/ })).toBeVisible({ timeout: 60_000 });
    const project = await findProject(api, name);
    expect(project, `project ${name} is made`).toBeDefined();
    const near: App = { id: "", name: "", projectId: project!.id, env: "prod" };
    const printer = await findAppNamed(api, near, "printer");
    const site = await findAppNamed(api, near, "site");
    expect(printer && site, "both services are apps").toBeTruthy();
    await deployed(api, printer!);
    await deployed(api, site!);

    // The secret where compose puts it, the file where it is mounted.
    await expectLogs(page, printer!, "TOKEN=token-from-the-folder");
    await expectLogs(page, printer!, "CONF=conf-from-the-folder");
    // The directory's file, served at the labels' host.
    await visit(page, `https://${domain}/`);
    await expect(page.locator("h1")).toHaveText("site-from-the-folder");
});
