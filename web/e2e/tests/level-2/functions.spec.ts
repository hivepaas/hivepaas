import { type App, createProject, deleteProject, findAppNamed } from "../../support/api";
import { appPage, deployed } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { domainFor, visit } from "../../support/routing";

// The first build pulls the runtime's images and starts buildkit: room for it.
test.describe.configure({ timeout: 600_000 });
// The proxy's certificate for a .localhost name is its own, signed by no one.
test.use({ ignoreHTTPSErrors: true });

test("a function made from the runtime's template builds, and answers a call, at its domain too", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("function"));
    cleanup(() => deleteProject(api, project.id));
    const domain = domainFor("function");

    await page.goto(`/projects/${project.id}/apps/`);
    await page.getByRole("button", { name: "New Function" }).click();
    const dialog = page.getByRole("dialog", { name: "Create Function" });
    await dialog.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("hello");
    await dialog.getByRole("checkbox", { name: "Expose at a domain" }).check();
    await dialog.getByRole("textbox", { name: "Domain" }).fill(domain);
    await dialog.getByRole("button", { name: "Create Function" }).click();

    const near: App = { id: "", name: "", projectId: project.id, env: "development" };
    await expect.poll(() => findAppNamed(api, near, "hello"), { timeout: 60_000 }).toBeDefined();
    const fn = (await findAppNamed(api, near, "hello"))!;
    await deployed(api, fn);

    await page.goto(appPage(fn, "code"));
    await page.getByRole("textbox", { name: "Path and query" }).fill("/hello?name=Ada");
    await page.getByRole("button", { name: "Run", exact: true }).click();
    // A run of its own, in a container that goes once it answers.
    await expect(page.getByRole("tabpanel", { name: "Body" })).toContainText('"hello": "Ada"', { timeout: 180_000 });
    await expect(page.locator("main")).toContainText(/OK\s*200/);

    // As a caller reaches it: at its domain, sent to HTTPS.
    await visit(page, `http://${domain}/hello?name=Grace`);
    expect(page.url()).toBe(`https://${domain}/hello?name=Grace`);
    await expect(page.locator("body")).toContainText(/"hello":\s*"Grace"/);
});

// PHP's template: its handler in index.php, its composer.json installed by
// Composer; run, then called at its domain. Each call runs in a process that
// keeps the handler loaded.
test("a PHP function made from its template builds with Composer, and answers a call, at its domain too", async ({
    page,
    api,
    cleanup,
}) => {
    const project = await createProject(api, e2eName("function-php"));
    cleanup(() => deleteProject(api, project.id));
    const domain = domainFor("function-php");

    await page.goto(`/projects/${project.id}/apps/`);
    await page.getByRole("button", { name: "New Function" }).click();
    const dialog = page.getByRole("dialog", { name: "Create Function" });
    await dialog.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("hello-php");
    await dialog.getByRole("group", { name: "Runtime *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: "PHP 8.5" }).click();
    await dialog.getByRole("checkbox", { name: "Expose at a domain" }).check();
    await dialog.getByRole("textbox", { name: "Domain" }).fill(domain);
    await dialog.getByRole("button", { name: "Create Function" }).click();

    const near: App = { id: "", name: "", projectId: project.id, env: "development" };
    await expect.poll(() => findAppNamed(api, near, "hello-php"), { timeout: 60_000 }).toBeDefined();
    const fn = (await findAppNamed(api, near, "hello-php"))!;
    await deployed(api, fn);

    await page.goto(appPage(fn, "code"));
    await expect(page.getByText("index.php", { exact: true }).first()).toBeVisible();
    await page.getByRole("textbox", { name: "Path and query" }).fill("/hello?name=Ada");
    await page.getByRole("button", { name: "Run", exact: true }).click();
    await expect(page.getByRole("tabpanel", { name: "Body" })).toContainText('"hello": "Ada"', { timeout: 180_000 });
    await expect(page.locator("main")).toContainText(/OK\s*200/);

    await visit(page, `http://${domain}/hello?name=Grace`);
    expect(page.url()).toBe(`https://${domain}/hello?name=Grace`);
    await expect(page.locator("body")).toContainText(/"hello":\s*"Grace"/);
});
