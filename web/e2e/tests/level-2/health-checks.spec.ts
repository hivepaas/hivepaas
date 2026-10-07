import type { APIRequestContext, Page } from "@playwright/test";

import { type App, appPath, deployImage } from "../../support/api";
import { WHOAMI, appIn, appPage, deployed } from "../../support/apps";
import { expect, test } from "../../support/fixtures";
import { domainFor, exposeAt, visit } from "../../support/routing";

test.describe.configure({ timeout: 300_000 });

// The checks are run by the backend, which in dind is on no app's network: the
// URL they call is the backend's own, or an app's domain through the proxy.
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

// expectRuns waits for the check's runs, as its row's View Runs shows them, to
// end as `status`, and finds none that ended as `never`.
async function expectRuns(
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

test("a health check runs on its interval: done when answered as asked, failed when not", async ({
    page,
    api,
    cleanup,
}) => {
    const app = await appIn(api, cleanup, "health");
    await newHealthCheck(page, app, "e2e-ping", `${BACKEND}/api/ping`);
    await newHealthCheck(page, app, "e2e-missing", `${BACKEND}/api/no-such-path`);

    await expectRuns(page, api, app, "e2e-ping", "Done", "Failed");
    await expectRuns(page, api, app, "e2e-missing", "Failed", "Done");
});

test("a health check of an app at its own domain passes", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "health-own");
    await deployImage(api, app, WHOAMI);
    await deployed(api, app);
    // Over plain HTTP: the proxy's certificate for a .localhost name is signed
    // by no one, and the check trusts only those that are.
    const domain = domainFor("health-own");
    await exposeAt(page, app, domain, async () => {
        await page.getByRole("group", { name: "Force HTTPS" }).getByRole("checkbox").uncheck();
    });
    // Routed first: the proxy takes a few seconds to read the app's labels.
    await visit(page, `http://${domain}/health`);

    await newHealthCheck(page, app, "e2e-own", `http://${domain}/health`);

    await expectRuns(page, api, app, "e2e-own", "Done", "Failed");
});
