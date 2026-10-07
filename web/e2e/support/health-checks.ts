import type { APIRequestContext, Page } from "@playwright/test";

import { type App, appPath } from "./api";
import { appPage } from "./apps";
import { expect } from "./fixtures";

// newHealthCheck makes a health check of the app from its Periodic Jobs: the
// URL called every 10 seconds, passing on a 200 as the form has it. `configure`
// fills more of the form before it is saved.
export async function newHealthCheck(
    page: Page,
    app: App,
    name: string,
    url: string,
    configure?: () => Promise<void>,
): Promise<void> {
    await page.goto(appPage(app, "periodic-jobs"));
    await page.getByRole("button", { name: "New Health Check" }).click();
    await page.getByRole("group", { name: "Name *" }).getByRole("textbox").fill(name);
    await page.getByRole("group", { name: "Interval *" }).getByRole("textbox").fill("10s");
    await page.getByRole("group", { name: "URL *" }).getByRole("textbox").fill(url);
    await configure?.();
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByRole("row", { name: new RegExp(name) })).toBeVisible();
}

// expectRuns waits for the check's runs, as its row's View Runs shows them, to
// end as `status`, and finds none that ended as `never`.
export async function expectRuns(
    page: Page,
    api: APIRequestContext,
    app: App,
    name: string,
    status: string,
    never: string,
): Promise<void> {
    const jobs = (await (
        await api.get(`${appPath(app)}/periodic-jobs`, { params: { kind: "healthcheck" } })
    ).json()) as {
        data: { id: string; name: string }[];
    };
    const id = jobs.data.find(job => job.name === name)?.id ?? "";
    await expect(async () => {
        await page.goto(`${appPage(app, "tasks")}?targetId=${id}`);
        await expect(
            page.getByRole("button", { name: new RegExp(`^${status} task:periodic-exec`) }).first(),
        ).toBeVisible({ timeout: 3_000 });
    }).toPass({ timeout: 60_000, intervals: [3_000] });
    await expect(page.getByRole("button", { name: new RegExp(`^${never} task:periodic-exec`) })).toHaveCount(0);
}
