import type { Page } from "@playwright/test";

import { type App, appPath } from "../../support/api";
import { appIn, appPage } from "../../support/apps";
import { expect, test } from "../../support/fixtures";

test.describe.configure({ timeout: 300_000 });

// The checks are run by the backend, which in dind is on no app's network: the
// URL they call is the backend's own.
const BACKEND = "http://localhost:10100";

async function newHealthCheck(page: Page, app: App, name: string, url: string): Promise<void> {
    await page.goto(appPage(app, "periodic-jobs"));
    await page.getByRole("button", { name: "New Health Check" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill(name);
    await page.getByRole("group", { name: "Interval *" }).getByRole("textbox").fill("10s");
    await page.getByRole("group", { name: "URL *" }).getByRole("textbox").fill(url);
    // Left as the form has it: a check passes on a 200.
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: new RegExp(name) })).toBeVisible();
}

test("a health check runs on its interval: done when answered as asked, failed when not", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "health");
    await newHealthCheck(page, app, "e2e-ping", `${BACKEND}/api/ping`);
    await newHealthCheck(page, app, "e2e-missing", `${BACKEND}/api/no-such-path`);

    const jobs = (await (
        await api.get(`${appPath(app)}/periodic-jobs`, { params: { kind: "healthcheck" } })
    ).json()) as {
        data: { id: string; name: string }[];
    };
    const idOf = (name: string) => jobs.data.find(job => job.name === name)?.id ?? "";
    for (const [name, status, never] of [
        ["e2e-ping", "Done", "Failed"],
        ["e2e-missing", "Failed", "Done"],
    ] as const) {
        // A check's runs, as its row's View Runs shows them.
        await expect(async () => {
            await page.goto(`${appPage(app, "tasks")}?targetId=${idOf(name)}`);
            await expect(
                page.getByRole("button", { name: new RegExp(`^${status} task:periodic-exec`) }).first(),
            ).toBeVisible({
                timeout: 3_000,
            });
        }).toPass({ timeout: 60_000, intervals: [3_000] });
        await expect(page.getByRole("button", { name: new RegExp(`^${never} task:periodic-exec`) })).toHaveCount(0);
    }
});
