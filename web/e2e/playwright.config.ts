import { defineConfig, devices } from "@playwright/test";

import { authFile, env, isLocalBackend } from "./support/env";

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

export default defineConfig({
    testDir: "tests",
    fullyParallel: true,
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
            testIgnore: levelZeroOnly ? /[\\/]level-[12][\\/]/ : undefined,
            use: {
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
            },
            dependencies: ["setup"],
        },
    ],
});
