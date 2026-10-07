import type { APIRequestContext, Page } from "@playwright/test";

import { type App, createApp, createProject, deleteProject, latestDeployment } from "./api";
import { e2eName, expect } from "./fixtures";

// What level 2 deploys: `whoami` for an app that serves, `busybox` for one that
// prints what it was given.
export const WHOAMI = "traefik/whoami:latest";
export const BUSYBOX = "busybox:1.37";

// A deployment pulls its image the first time: it has room.
export const DEPLOYED = { timeout: 180_000, intervals: [2_000] };

type Cleanup = (step: () => Promise<unknown>) => void;

// appIn makes a project of its own with one app in it; the project, and the
// app with it, go once the test ends.
export async function appIn(api: APIRequestContext, cleanup: Cleanup, label: string): Promise<App> {
    const project = await createProject(api, e2eName(label));
    cleanup(() => deleteProject(api, project.id));
    return createApp(api, project, "web");
}

// appPage is the address of one of the app's tabs.
export const appPage = (app: App, tab: string) => `/projects/${app.projectId}/${app.env}/apps/${app.id}/${tab}/`;

// deployed waits for the app's deployment - its newest, or the one given - to
// be done.
export async function deployed(api: APIRequestContext, app: App, id?: string): Promise<void> {
    await expect
        .poll(async () => {
            const latest = await latestDeployment(api, app);
            return id && latest?.id !== id ? "not yet" : latest?.status;
        }, DEPLOYED)
        .toBe("done");
}

// redeploy re-deploys the app from its header, and waits for that deployment
// to be done.
export async function redeploy(page: Page, api: APIRequestContext, app: App): Promise<void> {
    const before = await latestDeployment(api, app);
    await page.goto(appPage(app, "general"));
    await page.getByRole("button", { name: "Re-deploy" }).click();
    await page.getByRole("dialog", { name: "Re-deploy app" }).getByRole("button", { name: "Re-deploy" }).click();
    await expect.poll(async () => (await latestDeployment(api, app))?.id, DEPLOYED).not.toBe(before?.id);
    await deployed(api, app);
}

// restart restarts the app from its header: its containers are made anew, on
// the same spec. A re-deploy of an unchanged app changes nothing for swarm to
// act on.
export async function restart(page: Page, app: App): Promise<void> {
    await page.goto(appPage(app, "general"));
    await page.getByRole("button", { name: "Restart", exact: true }).click();
    await page.getByRole("dialog", { name: "Restart app" }).getByRole("button", { name: "Restart" }).click();
}

// expectInstances reads, after a reload, how many of the app's replicas run
// against how many are asked for, as the app's header says it: "1/1".
export async function expectInstances(page: Page, app: App, value: string): Promise<void> {
    await expect(async () => {
        await page.goto(appPage(app, "general"));
        await expect(page.getByRole("link", { name: /^Instances/ })).toContainText(value, { timeout: 2_000 });
    }).toPass({ timeout: 120_000, intervals: [3_000] });
}

// expectLogs waits for the app's log to hold the text.
export async function expectLogs(page: Page, app: App, text: string): Promise<void> {
    await page.goto(appPage(app, "logs"));
    await expectShownLogs(page, text);
}

// expectShownLogs waits for the log on the page - an app's, a task's - to hold
// the text. A log is drawn on a canvas: it is read the way a person can take
// it, Copy logs and the clipboard.
export async function expectShownLogs(page: Page, text: string): Promise<void> {
    await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
    await expect(async () => {
        await page.getByRole("button", { name: "Copy logs" }).locator("visible=true").first().click();
        const copied = await page.evaluate(() => navigator.clipboard.readText());
        expect(copied).toContain(text);
    }).toPass({ timeout: 60_000, intervals: [2_000] });
}
