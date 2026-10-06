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
            use: { ...devices["Desktop Chrome"], storageState: authFile },
            dependencies: ["setup"],
        },
    ],
});
