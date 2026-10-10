import { defineConfig, devices } from "@playwright/test";

import { authFile, env, isDevServer, isLocalBackend } from "./support/env";

// Levels 1 and 2 make, deploy and remove things: they run on a throwaway
// installation, never on the local backend, where only level 0 does. Said
// once, by the runner - its workers read this file too.
const levelZeroOnly = isLocalBackend(env.baseURL);
if (levelZeroOnly && process.env["TEST_WORKER_INDEX"] === undefined) {
    console.warn(
        `${env.baseURL} is the local backend: levels 1 and 2 are left out. They run on the throwaway ` +
            "installation - yarn env:up, then HP_E2E_BASE_URL=http://localhost:10100 (see the README).",
    );
}

// The tests that switch the server's return of secrets on: tests/**/secrets-on/.
const SECRETS_ON = /[\\/]secrets-on[\\/]/;

// The browser every test runs in.
const CHROME = {
    ...devices["Desktop Chrome"],
    storageState: authFile,
    launchOptions: {
        args: [
            // An app's domain, <name>.localhost, keeps its address - port
            // and all, as a redirect writes it - and is reached where the
            // proxy is.
            `--host-resolver-rules=MAP *.localhost:80 ${env.ingressHTTP}, MAP *.localhost:443 ${env.ingressHTTPS}`,
            // Chrome tries HTTPS first for an http:// address, which would
            // hide what the proxy does with plain HTTP.
            "--disable-features=HttpsUpgrades",
        ],
    },
};

export default defineConfig({
    testDir: "tests",
    fullyParallel: true,
    // The dev server builds the pages as they are opened: more at once than two,
    // and a page outwaits the assertion that it loaded.
    workers: isDevServer(env.baseURL) ? 2 : undefined,
    forbidOnly: !!process.env["CI"],
    retries: process.env["CI"] ? 1 : 0,
    reporter: [["list"], ["html", { open: "never" }]],
    use: {
        baseURL: env.baseURL,
        trace: "retain-on-failure",
        screenshot: "only-on-failure",
        video: "retain-on-failure",
    },
    projects: [
        // Signs in once, for every other test to start signed in.
        { name: "setup", testMatch: /.*\.setup\.ts/ },
        {
            name: "chromium",
            testIgnore: levelZeroOnly ? /[\\/]level-[12][\\/]/ : SECRETS_ON,
            use: CHROME,
            dependencies: ["setup"],
        },
        // The tests that need the server to return secrets through its API -
        // an export with them, for one - turn that on, and off again when they
        // end. Others count on it being off (2.26, 2.85), so these run alone,
        // once every other test is done - and a run of only them runs every
        // other test first. On their own, with what they depend on left out,
        // signed in by a run before on the same installation (--project=setup):
        // --project=secrets-on --no-deps.
        {
            name: "secrets-on",
            testMatch: SECRETS_ON,
            testIgnore: levelZeroOnly ? /.*/ : undefined,
            use: CHROME,
            dependencies: ["chromium"],
        },
    ],
});
