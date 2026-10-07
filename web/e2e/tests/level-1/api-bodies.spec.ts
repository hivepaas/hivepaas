import fs from "node:fs";
import path from "node:path";

import { createApp, createProject, deleteProject } from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";

// The endpoints are read from the backend's OpenAPI document, in its repo
// beside this one, or where HP_BACKEND_DIR says - as for env/up.sh.
const backendDir = process.env["HP_BACKEND_DIR"] ?? path.join(import.meta.dirname, "..", "..", "..", "..", "hivepaas");
const openAPIFile = path.join(backendDir, "docs", "openapi", "swagger.json");

// Left out, for what an empty body would still do there:
const SKIPPED = new RegExp(
    [
        // sign the tests' user out, or change who signs in and how;
        "logout|login|signup|invite|password|refresh|passcode|2fa|totp|mfa|sessions",
        // start work that goes on after the answer;
        "restart|exec|update-version|reload|/run|clone-execute|/deploy|redeploy|previews|from-compose",
        "import|spec|revert|confirm|cleanup|backup|test|/ai/",
        // deploy the system's own apps, or restart the proxy;
        "registry|logging|traefik|hivepaas/",
        // take a file, or answer a state;
        "photo|upload|download|running-status|status$",
        // change the installation's settings, which a run leaves as it found them.
        "settings/mcp|get-started",
    ].join("|"),
);

// An ID of the right shape that names nothing.
const NO_SUCH_ID = "01M4900000000000000000000Z";

interface OpenAPI {
    paths: Record<string, Record<string, unknown>>;
}

// A PUT or POST whose body names nothing is refused, or does what it does
// without one; it never fails on the server, and makes nothing. Such a body
// used to panic in 130 of them: the request types embed their fields' struct
// by pointer, and one no field filled was left nil.
test("an empty body is refused, or harmless, on every PUT and POST", async ({ api, cleanup }) => {
    test.skip(!fs.existsSync(openAPIFile), `no OpenAPI document at ${openAPIFile}: set HP_BACKEND_DIR`);
    test.setTimeout(180_000);

    const project = await createProject(api, e2eName("empty-bodies"));
    cleanup(() => deleteProject(api, project.id));
    const app = await createApp(api, project, "web");
    const params: Record<string, string> = { projectID: project.id, projectEnv: app.env, appID: app.id };

    const doc = JSON.parse(fs.readFileSync(openAPIFile, "utf8")) as OpenAPI;
    const failed: string[] = [];
    let tried = 0;
    for (const [route, methods] of Object.entries(doc.paths)) {
        if (SKIPPED.test(route)) continue;
        const url = route.replace(/^\//, "").replace(/\{(\w+)\}/g, (_, name: string) => params[name] ?? NO_SUCH_ID);
        for (const method of ["put", "post"] as const) {
            if (!(method in methods)) continue;
            tried++;
            const res = await api[method](url, { data: {} });
            if (res.status() >= 500 || res.status() === 201) {
                const body = (await res.text()).slice(0, 160);
                failed.push(`${method.toUpperCase()} ${route} -> ${res.status()} ${body}`);
            }
        }
    }

    expect(tried, "endpoints tried").toBeGreaterThan(100);
    expect(failed).toEqual([]);
});
