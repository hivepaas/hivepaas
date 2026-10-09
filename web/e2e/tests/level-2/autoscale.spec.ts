import type { APIRequestContext } from "@playwright/test";

import { type App, appPath } from "../../support/api";
import { appIn, appPage } from "../../support/apps";
import { expect, test } from "../../support/fixtures";

interface Autoscale {
    enabled: boolean;
    target: number;
    scaleInDelay: string;
    requestsUnavailable?: string;
    cpuUnavailable?: string;
    updateVer: number;
}

async function autoscaleOf(api: APIRequestContext, app: App): Promise<Autoscale> {
    const res = await api.get(`${appPath(app)}/autoscale`);
    expect(res.ok(), `reading the autoscale of ${app.name}: ${res.status()}`).toBe(true);
    return ((await res.json()) as { data: Autoscale }).data;
}

// An app reached by no domain, with no CPU limit, has neither of what autoscale
// scales on to read - its requests, its CPU - and an installation without its
// logs set up has neither either. Autoscale is not turned on then: the page
// says why before anything is sent, and the API refuses it as well, as it
// refuses a client that does not ask first - the CLI, a script.
test("autoscale is not turned on while what it scales on cannot be read", async ({ page, api, cleanup }) => {
    const app = await appIn(api, cleanup, "autoscale-unread");
    const before = await autoscaleOf(api, app);
    expect(before.requestsUnavailable, "its requests cannot be read").toBeTruthy();
    expect(before.cpuUnavailable, "its CPU cannot be read").toBeTruthy();

    const sent: string[] = [];
    page.on("request", req => {
        if (req.method() === "PUT" && req.url().endsWith("/autoscale")) sent.push(req.url());
    });
    await page.goto(appPage(app, "availability-and-scaling"));
    await page
        .getByRole("group", { name: /^Autoscale/ })
        .getByRole("checkbox")
        .check();
    await page.getByRole("button", { name: "Save Autoscale" }).click();

    await expect(page.getByText(/^What it scales on cannot be read now/)).toBeVisible();
    expect(sent, "nothing was sent").toEqual([]);

    const refused = await api.put(`${appPath(app)}/autoscale`, {
        data: {
            enabled: true,
            minReplicas: 1,
            maxReplicas: 3,
            target: before.target,
            requestsTarget: 0,
            cpuTarget: 70,
            scaleInDelay: before.scaleInDelay,
            updateVer: before.updateVer,
        },
    });
    expect(refused.status(), "the API refuses it").toBe(412);
    expect(((await refused.json()) as { code: string }).code).toBe("ERR_AUTOSCALE_UNREADABLE");
    expect((await autoscaleOf(api, app)).enabled, "autoscale stays off").toBe(false);
});
