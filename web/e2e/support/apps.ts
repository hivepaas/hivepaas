import type { APIRequestContext, Page } from "@playwright/test";

import { type App, appPath, createApp, createProject, deleteProject, latestDeployment } from "./api";
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

export interface MountOptions {
    // A directory below the app's own on the volume.
    subpath?: string;
    readOnly?: boolean;
    // Another app of the environment, by its name, whose directory the mount
    // reaches instead of the app's own; read only, unless allowWriting.
    dataOf?: string;
    allowWriting?: boolean;
}

// fillMount fills Persistent Storage's form: the volume, mounted at the target,
// as the options say.
export async function fillMount(page: Page, volume: string, target: string, options: MountOptions = {}): Promise<void> {
    await page.getByRole("group", { name: "Volume *" }).getByRole("combobox").click();
    await page.getByRole("option", { name: volume }).click();
    if (options.subpath) {
        await page.getByRole("group", { name: "Subpath", exact: true }).getByRole("textbox").fill(options.subpath);
    }
    if (options.dataOf) {
        await page
            .getByRole("group", { name: /^Data of/ })
            .getByRole("combobox")
            .click();
        await page.getByRole("option", { name: options.dataOf, exact: true }).click();
    }
    if (options.allowWriting) {
        await page.getByRole("group", { name: "Allow writing", exact: true }).getByRole("checkbox").check();
    }
    if (options.readOnly) {
        await page.getByRole("group", { name: "Read-only", exact: true }).getByRole("checkbox").check();
    }
    await page.getByRole("group", { name: "Target *" }).getByRole("textbox").fill(target);
}

// mountVolume mounts a volume in the app from Persistent Storage.
export async function mountVolume(
    page: Page,
    app: App,
    volume: string,
    target: string,
    options: MountOptions = {},
): Promise<void> {
    await page.goto(appPage(app, "persistent-storage"));
    await page.getByRole("button", { name: "New Storage Mount" }).click();
    await fillMount(page, volume, target, options);
    await page.getByRole("button", { name: "Save" }).click();
    // Saved, the list shows it: a busy node takes a while to check the storage
    // and update the service.
    await expect(mountRow(page, target)).toBeVisible({ timeout: 30_000 });
}

// mountRow is Persistent Storage's row of the mount at the target.
export const mountRow = (page: Page, target: string) =>
    page.getByRole("row").filter({ has: page.getByRole("cell", { name: target, exact: true }) });

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

// expectLogs waits for the app's log to hold the text, or a match of the
// pattern, and answers what it found. The log is the app's current
// container's: one replaced takes its log with it.
export async function expectLogs(page: Page, app: App, text: string | RegExp): Promise<string> {
    await page.goto(appPage(app, "logs"));
    return expectShownLogs(page, text);
}

// expectShownLogs waits for the log on the page - an app's, a task's - to hold
// the text, or a match of the pattern, and answers what it found. A log is
// drawn on a canvas: it is read the way a person can take it, Copy logs and
// the clipboard.
export async function expectShownLogs(page: Page, text: string | RegExp): Promise<string> {
    let found = "";
    await expect(async () => {
        const copied = await copyShownLogs(page);
        const match = typeof text === "string" ? (copied.includes(text) ? text : null) : copied.match(text)?.[0];
        expect(match, `${String(text)} in the log`).toBeTruthy();
        found = match ?? "";
    }).toPass({ timeout: 60_000, intervals: [2_000] });
    return found;
}

// copyShownLogs is the log on the page, as Copy logs puts it on the clipboard.
export async function copyShownLogs(page: Page): Promise<string> {
    await page.context().grantPermissions(["clipboard-read", "clipboard-write"]);
    await page.getByRole("button", { name: "Copy logs" }).locator("visible=true").first().click();
    return page.evaluate(() => navigator.clipboard.readText());
}

interface ServiceTask {
    id: string;
    desiredState: string;
    status: { state: string };
}

// runningTask is the app's container once swarm has settled on it: the only
// one meant to run, and running.
export async function runningTask(api: APIRequestContext, app: App): Promise<string> {
    let id = "";
    await expect
        .poll(
            async () => {
                const body = (await (await api.get(`${appPath(app)}/service-tasks`)).json()) as {
                    data: ServiceTask[];
                };
                const meant = body.data.filter(task => task.desiredState === "running");
                id = meant.length === 1 && meant[0]?.status.state === "running" ? meant[0].id : "";
                return id;
            },
            { timeout: 60_000 },
        )
        .not.toBe("");
    return id;
}

// FOLLOWED is how long a running app takes to follow a change it is not
// deployed for - a variable, a secret, a mounted file: its service is updated,
// and swarm starts a new container, on an image it has.
export const FOLLOWED = { timeout: 120_000, intervals: [2_000] };

// lastPrinted is the value the app's container printed last for the key, as
// key=value, in the log on the page: the value of the newest container, once
// the one before is gone.
export async function lastPrinted(page: Page, key: string): Promise<string> {
    const values = [...(await copyShownLogs(page)).matchAll(new RegExp(`\\b${key}=(\\S*)`, "g"))];
    return values.at(-1)?.[1] ?? "";
}

// expectPrinted waits for the last value the app printed for the key to be
// value, or to pass the check, and answers it. The app prints it over and over:
// what a change it follows does is read in the newest container's log.
export async function expectPrinted(
    page: Page,
    app: App,
    key: string,
    value: string | ((printed: string) => boolean),
): Promise<string> {
    await page.goto(appPage(app, "logs"));
    const check = typeof value === "string" ? (printed: string) => printed === value : value;
    let printed = "";
    await expect
        .poll(async () => {
            printed = await lastPrinted(page, key);
            return check(printed) ? "yes" : printed;
        }, FOLLOWED)
        .toBe("yes");
    return printed;
}
