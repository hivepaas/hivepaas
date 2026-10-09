import type { Page, Response } from "@playwright/test";

import type { App } from "./api";
import { appPage } from "./apps";
import { e2eName, expect } from "./fixtures";

// domainFor names a domain for what a test exposes: <its e2e name>.localhost,
// which the browser reaches at the installation's proxy (playwright.config.ts).
export const domainFor = (label: string) => `${e2eName(label)}.localhost`;

// exposeAt opens the app to the internet at the domain, on the container's
// port 80, from Routing Settings. `configure` fills more of the domain's
// settings before they are saved.
export async function exposeAt(page: Page, app: App, domain: string, configure?: () => Promise<void>): Promise<void> {
    await page.goto(appPage(app, "routing-settings"));
    await page.getByRole("group", { name: "Container Port" }).getByRole("textbox").fill("80");
    await page.getByRole("group", { name: "Expose The App To The Internet" }).getByRole("checkbox").check();
    await addDomain(page, domain);
    await configure?.();
    await page.getByRole("button", { name: "Save" }).click();
    await expect(page.getByText("Routing settings updated")).toBeVisible();
}

// addDomain adds a domain to the app's, and selects it: the settings below are
// the new domain's.
export async function addDomain(page: Page, domain: string): Promise<void> {
    const domains = page.getByRole("group", { name: "Domains" });
    await domains.getByRole("button", { name: "Add domain" }).click();
    await domains.getByRole("textbox").fill(domain);
    await domains.getByRole("button", { name: "Confirm" }).click();
    await domains.getByRole("button", { name: domain, exact: true }).click();
    await expect(page.getByText(`Selected domain: ${domain}`)).toBeVisible();
}

// addSection adds one of the domain's optional settings, by its menu name.
export async function addSection(page: Page, name: string): Promise<void> {
    await page.getByRole("button", { name: "Add Configuration" }).click();
    await page.getByRole("menuitem", { name }).click();
}

// visit opens the address until the proxy answers it as asked, and answers
// that response. A domain just set takes a while: the proxy reads the apps'
// labels every 15 seconds, and the app's container is made anew to join the
// proxy's network - on a busy node, a minute or more.
export async function visit(page: Page, url: string, status = 200): Promise<Response> {
    let res: Response | null = null;
    await expect(async () => {
        res = await page.goto(url);
        expect(res?.status(), `${url} answers ${status}`).toBe(status);
    }).toPass({ timeout: 120_000, intervals: [2_000] });
    return res as unknown as Response;
}
