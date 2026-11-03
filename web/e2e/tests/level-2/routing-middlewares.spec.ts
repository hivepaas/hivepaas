import type { Page } from "@playwright/test";

import { deployImage } from "../../support/api";
import { BUSYBOX, appIn, appPage, deployed } from "../../support/apps";
import { expect, test } from "../../support/fixtures";
import {
    type Fetched,
    addDomain,
    addSection,
    domainFor,
    exposeAt,
    fetchFrom,
    visit,
    whoamiIn,
} from "../../support/routing";

// What a domain's optional settings do to the requests that reach it, and to
// the answers: each test sets them in Routing Settings, as a person does, and
// asks the domain through the proxy. The app is whoami, which answers with what
// reached it.

// Each test deploys, then waits for the proxy to route what it set: room for it.
test.describe.configure({ timeout: 240_000 });
// The proxy's certificate for a .localhost name is its own, signed by no one.
test.use({ ignoreHTTPSErrors: true });

// group is a block of the selected domain's settings, by its title.
const group = (page: Page, title: string) => page.getByRole("group", { name: title, exact: true });

// rewriteMode picks how the path is rewritten: Add Prefix, Strip Prefix or
// Replace.
const rewriteMode = (page: Page, name: string) => group(page, "Path Rewrite Mode").getByRole("tab", { name }).click();

// answered opens the domain, and waits for the proxy to route it: what the page
// fetches after is asked of the domain.
async function answered(page: Page, domain: string): Promise<void> {
    await visit(page, `https://${domain}/`);
}

// answerOf is whoami's answer to a request with a header of `size` bytes, which
// it echoes. Its other answers fit no size: /data writes as it goes, and the
// proxy compresses what is written in pieces, whatever its size.
async function answerOf(page: Page, size: number): Promise<Fetched> {
    return page.evaluate(async size => {
        const res = await fetch("/", { headers: { "X-Pad": "p".repeat(size) }, cache: "no-store" });
        return { status: res.status, headers: Object.fromEntries(res.headers.entries()), body: await res.text() };
    }, size);
}

test("a response is compressed for a client that accepts it, unless it is small or of a type left out", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-zip");
    const domain = domainFor("route-zip");
    const plain = domainFor("route-zip-plain");

    await exposeAt(page, app, domain, async () => {
        // An app's first domain compresses from the start, past 1 KB.
        await page.getByRole("button", { name: /^Compression Configuration/ }).click();
        await expect(group(page, "Enabled").getByRole("checkbox")).toBeChecked();
        const minimum = group(page, "Min Response Body Size").getByRole("textbox");
        await expect(minimum).toHaveValue("1kb");
        await minimum.fill("2kb");

        await addDomain(page, plain);
        await addSection(page, "Compression Configuration");
        await group(page, "Excluded Content Types").getByRole("textbox").fill("text/plain");
    });

    // whoami's answer echoes the request's headers: a header of the size asked
    // for makes it that much longer. The browser accepts zstd, br and gzip; the
    // proxy picks the first of its own it accepts.
    await answered(page, domain);
    const large = await answerOf(page, 4000);
    expect(large.headers["content-encoding"], "about 5 KB of text is compressed").toMatch(/^(zstd|br|gzip)$/);
    expect(large.body).toContain("p".repeat(4000));
    const middling = await answerOf(page, 600);
    expect(middling.headers["content-encoding"], "about 1.5 KB, under the 2 KB set, is not").toBeUndefined();

    await answered(page, plain);
    const excluded = await answerOf(page, 4000);
    expect(excluded.headers["content-encoding"], "text/plain left out is not").toBeUndefined();
});

test("the circuit breaker answers for the app while it fails, then lets requests through again", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-breaker");
    const domain = domainFor("route-breaker");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Circuit Breaker Configuration");
        await group(page, "Expression").getByRole("textbox").fill("ResponseCodeRatio(500, 600, 0, 600) > 0.5");
        await group(page, "Check Period").getByRole("textbox").fill("100ms");
        await group(page, "Fallback Duration").getByRole("textbox").fill("15s");
        await group(page, "Recovery Duration").getByRole("textbox").fill("5s");
        // A code no one else answers with: it can only be the breaker's.
        await group(page, "Response Code").getByRole("textbox").fill("418");
    });

    await answered(page, domain);
    // whoami's /health fails from now on; asked enough, the breaker opens, and
    // answers even what the app would answer well.
    expect((await fetchFrom(page, "/health", { method: "POST", body: "500" })).status).toBe(200);
    await expect(async () => {
        for (let i = 0; i < 5; i++) await fetchFrom(page, "/health");
        expect((await fetchFrom(page, "/")).status, "the breaker answers /").toBe(418);
    }).toPass({ timeout: 30_000 });

    // Once the fallback is over, and the failures out of its window, the app
    // answers again.
    await expect(async () => {
        expect((await fetchFrom(page, "/")).status, "the app answers /").toBe(200);
    }).toPass({ timeout: 90_000, intervals: [2_000] });
});

test("a client within the allowed IPs is let in, and a body over the size allowed is refused", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-client");
    const open = domainFor("route-client-open");
    const domain = domainFor("route-client");

    // The address the proxy sees this browser come from, as the app is told.
    await exposeAt(page, app, open, async () => {
        await page.getByRole("group", { name: "Force HTTPS" }).getByRole("checkbox").uncheck();
    });
    await visit(page, `http://${open}/`);
    const ip = /X-Real-Ip: (\S+)/.exec((await fetchFrom(page, "/")).body)?.[1];
    expect(ip, "whoami says where the request came from").toBeTruthy();

    await page.goto(appPage(app, "routing-settings"));
    await addDomain(page, domain);
    await addSection(page, "Client Configuration");
    await group(page, "Allowed IPs").getByRole("textbox").fill(`${ip}/32`);
    await group(page, "Max Request Body Size").getByRole("textbox").fill("1kb");
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Routing settings updated")).toBeVisible();

    await answered(page, domain);
    expect((await fetchFrom(page, "/", { method: "POST", body: "x".repeat(500) })).status, "500 bytes").toBe(200);
    expect((await fetchFrom(page, "/", { method: "POST", body: "x".repeat(4096) })).status, "4 KB").toBe(413);
});

// A CGI script of busybox's web server: it says which language the request
// asked for, with a header of its own, X-Powered-By; /cgi-bin/untyped says
// nothing of its content's type.
const CGI_SERVER = [
    "mkdir -p /www/cgi-bin",
    `printf '#!/bin/sh\\necho "X-Powered-By: busybox-cgi"\\necho "Content-Type: text/plain"\\necho\\necho "lang=$HTTP_ACCEPT_LANGUAGE"\\n' >/www/cgi-bin/typed`,
    `printf '#!/bin/sh\\necho\\necho untyped text\\n' >/www/cgi-bin/untyped`,
    "chmod +x /www/cgi-bin/*",
    "exec httpd -f -p 80 -h /www",
].join(" && ");

test("headers are removed from requests and responses, and a response's type is detected", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "route-unset");
    await deployImage(api, app, BUSYBOX, `sh -c '${CGI_SERVER.replace(/'/g, "'\\''")}'`);
    await deployed(api, app);
    const removing = domainFor("route-unset");
    const detecting = domainFor("route-detect");
    const untouched = domainFor("route-untouched");

    await exposeAt(page, app, removing, async () => {
        await addSection(page, "Header Configuration");
        const requests = page.getByRole("group", { name: "To Remove From Requests" });
        await requests.getByRole("textbox").fill("Accept-Language");
        await requests.getByRole("button", { name: "Add" }).click();
        const responses = page.getByRole("group", { name: "To Remove From Responses" });
        await responses.getByRole("textbox").fill("X-Powered-By");
        await responses.getByRole("button", { name: "Add" }).click();

        await addDomain(page, detecting);
        await addSection(page, "Header Configuration");
        await group(page, "Auto Detect Content Type").getByRole("checkbox").check();

        await addDomain(page, untouched);
    });

    await visit(page, `https://${untouched}/cgi-bin/typed`);
    const as = await fetchFrom(page, "/cgi-bin/typed");
    expect(as.body, "untouched, the app reads the browser's language").toContain("lang=en-US");
    expect(as.headers["x-powered-by"]).toBe("busybox-cgi");
    expect((await fetchFrom(page, "/cgi-bin/untyped")).headers["content-type"], "nor typed").toBeUndefined();

    await visit(page, `https://${removing}/cgi-bin/typed`);
    const removed = await fetchFrom(page, "/cgi-bin/typed");
    expect(removed.body, "the request reaches the app without its Accept-Language").toContain("lang=\n");
    expect(removed.headers["x-powered-by"], "the answer comes without X-Powered-By").toBeUndefined();

    await visit(page, `https://${detecting}/cgi-bin/typed`);
    const detected = await fetchFrom(page, "/cgi-bin/untyped");
    expect(detected.headers["content-type"], "the proxy names the type it finds").toBe("text/plain; charset=utf-8");
});

test("a path is given a prefix, stripped of one by a pattern, or replaced by a pattern", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-paths");
    const prefixed = domainFor("route-prefix");
    const stripped = domainFor("route-strip");
    const patterned = domainFor("route-pattern");

    await exposeAt(page, app, prefixed, async () => {
        await addSection(page, "Path Rewrite Configuration");
        await rewriteMode(page, "Add Prefix");
        await group(page, "Add Prefix").getByRole("textbox").fill("/v2");

        await addDomain(page, stripped);
        await addSection(page, "Path Rewrite Configuration");
        await rewriteMode(page, "Strip Prefix");
        await group(page, "Strip Prefix").getByRole("textbox").fill("/v[0-9]+");
        await group(page, "Strip Prefix").getByRole("checkbox", { name: "Is Regex" }).check();

        await addDomain(page, patterned);
        await addSection(page, "Path Rewrite Configuration");
        await rewriteMode(page, "Replace");
        await group(page, "Replace Path").getByRole("textbox").fill("^/blog/([0-9]+)$");
        await group(page, "Replace Path").getByRole("checkbox", { name: "Is Regex" }).check();
        await group(page, "Replace Path With").getByRole("textbox").fill("/posts/$1/view");
    });

    await answered(page, prefixed);
    expect((await fetchFrom(page, "/hello")).body).toContain("GET /v2/hello HTTP/1.1");

    await answered(page, stripped);
    expect((await fetchFrom(page, "/v3/items")).body).toContain("GET /items HTTP/1.1");

    await answered(page, patterned);
    expect((await fetchFrom(page, "/blog/42")).body).toContain("GET /posts/42/view HTTP/1.1");
});

// A path written out - no pattern - is replaced, and what is under it; a path
// that only starts the same is not.
test("a path is replaced with another, and what is under it", async ({ page, api, cleanup }) => {
    const app = await whoamiIn(api, cleanup, "route-replace");
    const domain = domainFor("route-replace");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Path Rewrite Configuration");
        await rewriteMode(page, "Replace");
        await group(page, "Replace Path").getByRole("textbox").fill("/old");
        await group(page, "Replace Path With").getByRole("textbox").fill("/new");
    });

    await answered(page, domain);
    expect((await fetchFrom(page, "/old")).body, "/old is replaced with /new").toContain("GET /new HTTP/1.1");
    expect((await fetchFrom(page, "/old/page")).body).toContain("GET /new/page HTTP/1.1");
    expect((await fetchFrom(page, "/older")).body, "/older is left alone").toContain("GET /older HTTP/1.1");
    expect((await fetchFrom(page, "/other")).body).toContain("GET /other HTTP/1.1");
});

test("requests beyond the in-flight limit are turned away while the others are answered", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await whoamiIn(api, cleanup, "route-inflight");
    const domain = domainFor("route-inflight");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Rate Limit Configuration");
        await group(page, "Max In-Flight Requests").getByRole("textbox").fill("1");
    });

    await answered(page, domain);
    // Two at once, each answered three seconds late: one is let through.
    const statuses = await page.evaluate(async () =>
        Promise.all(["/?wait=3s", "/?wait=3s"].map(async path => (await fetch(path, { cache: "no-store" })).status)),
    );
    expect(statuses.sort()).toEqual([200, 429]);
    expect((await fetchFrom(page, "/")).status, "one at a time is answered").toBe(200);
});

test("a setting turned off is kept, and not applied", async ({ page, api, cleanup }) => {
    const app = await whoamiIn(api, cleanup, "route-off");
    const domain = domainFor("route-off");

    await exposeAt(page, app, domain, async () => {
        await addSection(page, "Client Configuration");
        await group(page, "Allowed IPs").getByRole("textbox").fill("10.255.255.1/32");
        await group(page, "Enabled").getByRole("checkbox").uncheck();
    });

    // Off, the allowlist turns no one away.
    await answered(page, domain);

    await page.goto(appPage(app, "routing-settings"));
    await page.getByRole("group", { name: "Domains" }).getByRole("button", { name: domain, exact: true }).click();
    await page.getByRole("button", { name: /^Client Configuration/ }).click();
    await expect(group(page, "Enabled").getByRole("checkbox")).not.toBeChecked();
    await group(page, "Enabled").getByRole("checkbox").check();
    await expect(group(page, "Allowed IPs").getByRole("textbox")).toHaveValue("10.255.255.1/32");
});
