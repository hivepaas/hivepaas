import type { APIRequestContext, Browser } from "@playwright/test";

import { type App, appPath, createBasicAuth, deleteSettingsNamed, settingIdNamed } from "../../support/api";
import { appPage, expectInstances } from "../../support/apps";
import { e2eName, expect, test } from "../../support/fixtures";
import { domainFor, whoamiIn } from "../../support/routing";

// Each test deploys, then waits for the proxy to route what it set: room for it.
test.describe.configure({ timeout: 240_000 });
// The proxy's certificate for a .localhost name is its own, signed by no one.
test.use({ ignoreHTTPSErrors: true });

// routeAt exposes the app at the domain, over HTTPS only, with the domain's
// settings given - its paths', its load balancer's - as the API takes them.
async function routeAt(api: APIRequestContext, app: App, domain: string, settings: object): Promise<void> {
    const current = await api.get(`${appPath(app)}/routing-settings`);
    expect(current.ok(), `reading routing settings: ${current.status()}`).toBe(true);
    const { updateVer } = ((await current.json()) as { data: { updateVer: number } }).data;
    const res = await api.put(`${appPath(app)}/routing-settings`, {
        data: {
            port: 80,
            exposePublicly: true,
            domains: [{ enabled: true, domain, protocol: "http", forceHttps: true, ...settings }],
            updateVer,
        },
    });
    expect(res.ok(), `routing ${app.name} at ${domain}: ${res.status()} ${await res.text()}`).toBe(true);
}

interface Answer {
    status: number;
    body: string;
}

// ask opens the URL in a context of its own, with the credentials given, and
// answers what came back: whoami's answer is the request as it reached it. The
// browser reaches a .localhost domain where the proxy answers (see
// playwright.config.ts).
async function ask(
    browser: Browser,
    url: string,
    credentials?: { username: string; password: string },
): Promise<Answer> {
    const context = await browser.newContext({ ignoreHTTPSErrors: true, httpCredentials: credentials });
    try {
        const page = await context.newPage();
        const res = await page.goto(url);
        return { status: res?.status() ?? 0, body: (await res?.text()) ?? "" };
    } finally {
        await context.close();
    }
}

// A domain's paths have settings of their own, on top of the domain's: a path
// behind basic auth asks for it, the rest of the domain does not; a header is
// added to an exact path and not below it, to a pattern's paths and not
// others; and a path turned off changes nothing.
test("a domain's paths have settings of their own, matched exactly, by prefix or by a pattern", async ({
    page,
    browser,
    api,
    cleanup,
}) => {
    const credentials = e2eName("path-auth");
    await createBasicAuth(api, credentials, "visitor", "e2e-path-secret");
    cleanup(() => deleteSettingsNamed(api, "basic-auth", credentials));
    const authId = await settingIdNamed(api, "settings/basic-auth", credentials);
    const app = await whoamiIn(api, cleanup, "paths");
    const domain = domainFor("paths");
    const tagged = (value: string) => ({ enabled: true, toAddToRequests: { "X-E2e-Path": value } });
    await routeAt(api, app, domain, {
        paths: [
            { enabled: true, path: "/admin", mode: "prefix", basicAuth: { enabled: true, id: authId } },
            { enabled: true, path: "/exact", mode: "exact", headerConfig: tagged("exact") },
            { enabled: true, path: "^/v[0-9]+/", mode: "regex", headerConfig: tagged("pattern") },
            { enabled: false, path: "/off", mode: "prefix", basicAuth: { enabled: true, id: authId } },
        ],
    });

    const base = `https://${domain}`;
    await expect.poll(async () => (await ask(browser, `${base}/admin/users`)).status, { timeout: 120_000 }).toBe(401);
    const visitor = { username: "visitor", password: "e2e-path-secret" };
    expect((await ask(browser, `${base}/admin/users`, visitor)).status, "/admin/users, signed in").toBe(200);

    const answers = {
        "/": await ask(browser, `${base}/`),
        "/exact": await ask(browser, `${base}/exact`),
        "/exact/below": await ask(browser, `${base}/exact/below`),
        "/v2/items": await ask(browser, `${base}/v2/items`),
        "/vx/items": await ask(browser, `${base}/vx/items`),
        "/off/page": await ask(browser, `${base}/off/page`),
    };
    for (const [path, answer] of Object.entries(answers)) {
        expect.soft(answer.status, `${path} answers`).toBe(200);
    }
    expect.soft(answers["/"].body).not.toContain("X-E2e-Path");
    expect.soft(answers["/exact"].body).toContain("X-E2e-Path: exact");
    expect.soft(answers["/exact/below"].body).not.toContain("X-E2e-Path");
    expect.soft(answers["/v2/items"].body).toContain("X-E2e-Path: pattern");
    expect.soft(answers["/vx/items"].body).not.toContain("X-E2e-Path");

    // The paths are listed with their modes, as saved.
    await page.goto(appPage(app, "routing-settings"));
    await page.getByRole("group", { name: "Domains" }).getByRole("button", { name: domain, exact: true }).click();
    for (const [path, mode] of [
        ["/admin", "prefix"],
        ["/exact", "exact"],
        ["^/v[0-9]+/", "regex"],
        ["/off", "prefix"],
    ]) {
        await expect(page.getByRole("button", { name: `Path: ${path} ${mode}` })).toBeVisible();
    }
});

// replicasOf asks the domain the number of times given, and answers the
// replicas that answered, by their addresses: they have one hostname, the
// app's.
async function replicasOf(browser: Browser, url: string, times: number): Promise<Set<string>> {
    const replicas = new Set<string>();
    for (let i = 0; i < times; i++) {
        const answer = await ask(browser, url);
        const addresses = [...answer.body.matchAll(/^IP: (\S+)$/gm)].map(m => m[1]).filter(ip => ip !== "127.0.0.1");
        if (addresses.length > 0) replicas.add(addresses.sort().join(","));
    }
    return replicas;
}

// The proxy spreads requests over an app's replicas as its load balancer's
// strategy says: round robin to each in turn, highest random weight to the one
// a client is hashed to, every time.
test("the load balancer's strategy decides which replicas answer", async ({ page, browser, api, cleanup }) => {
    const app = await whoamiIn(api, cleanup, "balanced");
    await page.goto(appPage(app, "availability-and-scaling"));
    await page.getByRole("group", { name: "Replicas" }).getByRole("textbox").fill("2");
    await page.getByRole("button", { name: "Save", exact: true }).click();
    await expectInstances(page, app, "2/2");
    const domain = domainFor("balanced");
    const url = `https://${domain}/`;

    await routeAt(api, app, domain, { lbConfig: { strategy: "wrr" } });
    await expect.poll(async () => (await replicasOf(browser, url, 6)).size, { timeout: 120_000 }).toBe(2);

    await routeAt(api, app, domain, { lbConfig: { strategy: "hrw" } });
    // Once the proxy has the strategy, one replica answers each request.
    await expect.poll(async () => (await replicasOf(browser, url, 10)).size, { timeout: 120_000 }).toBe(1);
});

// A websocket opened at an app's domain reaches the app, through the proxy:
// whoami's /echo sends back what it is sent.
test("a websocket opened at the app's domain reaches the app", async ({ page, api, cleanup }) => {
    const app = await whoamiIn(api, cleanup, "sockets");
    const domain = domainFor("sockets");
    await routeAt(api, app, domain, {});
    await expect.poll(async () => (await page.goto(`https://${domain}/`))?.status(), { timeout: 120_000 }).toBe(200);

    const echoed = await page.evaluate(async (url: string) => {
        return new Promise<string>((resolve, reject) => {
            const socket = new WebSocket(url);
            socket.onopen = () => {
                socket.send("hello through the proxy");
            };
            socket.onmessage = event => {
                resolve(String(event.data));
                socket.close();
            };
            socket.onerror = () => {
                reject(new Error(`the websocket to ${url} failed`));
            };
        });
    }, `wss://${domain}/echo`);
    expect(echoed).toBe("hello through the proxy");
});
