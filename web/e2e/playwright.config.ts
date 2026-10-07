import { defineConfig, devices } from "@playwright/test";

import { authFile, env } from "./support/env";

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
