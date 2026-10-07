import type { APIRequestContext } from "@playwright/test";

import { type App, createBasicAuth, deleteSettingsNamed, deployImage } from "../../support/api";
import { WHOAMI, appIn, deployed } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { addDomain, addSection, domainFor, exposeAt, visit } from "../../support/routing";

// The proxy's certificate for a .localhost name is its own, signed by no one.
test.use({ ignoreHTTPSErrors: true });

type Cleanup = (step: () => Promise<unknown>) => void;

// whoamiIn is an app that answers each request with what reached it: the
// request line, the Host, the headers - what the proxy made of the request.
async function whoamiIn(api: APIRequestContext, cleanup: Cleanup, label: string): Promise<App> {
    const app = await appIn(api, cleanup, label);
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    return app;
}

test("an app exposed at a domain answers there over HTTPS, and Force HTTPS sends HTTP there", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-https");
    const domain = domainFor("route-https");

    await exposeAt(page, app, domain, async () => {
        // A new domain forces HTTPS.
        await expect(page.getByRole("group", { name: "Force HTTPS" }).getByRole("checkbox")).toBeChecked();
    });

    const res = await visit(page, `http://${domain}/hello`);
    expect(res.request().redirectedFrom()?.url()).toBe(`http://${domain}/hello`);
    expect(page.url()).toBe(`https://${domain}/hello`);
    await expect(page.locator("body")).toContainText("GET /hello HTTP/1.1");
    await expect(page.locator("body")).toContainText(`Host: ${domain}`);
});

test("without Force HTTPS, the app answers HTTP as well", async ({ page, api, cleanup }) => {
    const app = await whoamiIn(api, cleanup, "route-http");
    const domain = domainFor("route-http");

    await exposeAt(page, app, domain, async () => {
        await page.getByRole("group", { name: "Force HTTPS" }).getByRole("checkbox").uncheck();
    });

    await visit(page, `https://${domain}/`);
    const res = await visit(page, `http://${domain}/plain`);
    expect(res.request().redirectedFrom()).toBeNull();
    expect(page.url()).toBe(`http://${domain}/plain`);
    await expect(page.locator("body")).toContainText("GET /plain HTTP/1.1");
});

test("basic auth on a domain turns away a visitor without the credentials, and lets one with them in", async ({
    page,
    browser,
    api,
    cleanup,
}) => {
    // Made before the app, to be removed after it: the domain uses them.
    const credentials = e2eName("route-auth");
    await createBasicAuth(api, credentials, "visitor", "e2e-route-secret");
    cleanup(() => deleteSettingsNamed(api, "basic-auth", credentials));
    const app = await whoamiIn(api, cleanup, "route-auth");
    const domain = domainFor("route-auth");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Basic Auth");
        await page.getByRole("group", { name: "Basic Auth" }).getByRole("combobox").click();
        await page.getByRole("option", { name: credentials }).click();
    });

    await visit(page, `https://${domain}/`, 401);
    const visitor = await browser.newContext({
        ignoreHTTPSErrors: true,
        httpCredentials: { username: "visitor", password: "e2e-route-secret" },
    });
    try {
        const res = await (await visitor.newPage()).goto(`https://${domain}/`);
        expect(res?.status()).toBe(200);
    } finally {
        await visitor.close();
    }
});

test("a domain set to redirect sends its visitors to the main one, path and query kept", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-redirect");
    const main = domainFor("route-main");
    const old = domainFor("route-old");

    await exposeAt(page, app, main, async () => {
        await addDomain(page, old);
        await page.getByRole("group", { name: "Redirect To" }).getByRole("textbox").fill(main);
    });

    await visit(page, `https://${old}/docs?page=2`);
    expect(page.url()).toBe(`https://${main}/docs?page=2`);
});

test("headers are added on the way in and on the way out, and a path prefix is stripped", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-rewrite");
    const domain = domainFor("route-rewrite");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Header Configuration");
        const inbound = page.getByRole("group", { name: "To Add To Requests" });
        await inbound.getByRole("textbox", { name: "Name" }).fill("X-Asked-Through");
        await inbound.getByRole("textbox", { name: "Value" }).fill("hivepaas");
        await inbound.getByRole("button", { name: "Add" }).click();
        const outbound = page.getByRole("group", { name: "To Add To Responses" });
        await outbound.getByRole("textbox", { name: "Name" }).fill("X-Served-Through");
        await outbound.getByRole("textbox", { name: "Value" }).fill("hivepaas");
        await outbound.getByRole("button", { name: "Add" }).click();

        await addSection(page, "Path Rewrite Configuration");
        await page.getByRole("group", { name: "Path Rewrite Mode" }).getByRole("tab", { name: "Strip Prefix" }).click();
        await page.getByRole("group", { name: "Strip Prefix" }).getByRole("textbox").fill("/api");
    });

    const res = await visit(page, `https://${domain}/api/hello`);
    expect(res.headers()["x-served-through"]).toBe("hivepaas");
    await expect(page.locator("body")).toContainText("GET /hello HTTP/1.1");
    await expect(page.locator("body")).toContainText("X-Asked-Through: hivepaas");
});

test("a client outside the allowed IPs is turned away", async ({ page, api, cleanup }) => {
    const app = await whoamiIn(api, cleanup, "route-ips");
    const domain = domainFor("route-ips");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Client Configuration");
        await page.getByRole("group", { name: "Allowed IPs" }).getByRole("textbox").fill("10.255.255.1/32");
    });

    await visit(page, `https://${domain}/`, 403);
});

test("requests over the rate limit are turned away", async ({ page, api, cleanup }) => {
    const app = await whoamiIn(api, cleanup, "route-rate");
    const domain = domainFor("route-rate");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Rate Limit Configuration");
        await page.getByRole("group", { name: "Average" }).getByRole("textbox").fill("1");
        await page.getByRole("group", { name: "Period" }).getByRole("textbox").fill("1m");
        await page.getByRole("group", { name: "Burst" }).getByRole("textbox").fill("1");
    });

    // One a minute: the first is let through, the next is not.
    await visit(page, `https://${domain}/`);
    const again = await page.goto(`https://${domain}/`);
    expect(again?.status()).toBe(429);
});
