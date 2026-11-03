import type { APIRequestContext, Page } from "@playwright/test";

import { createApp, createProject, deleteProject, deleteSettingsNamed, findAppNamed, ok } from "../../../support/api";
import { e2eName, expect, test } from "../../../support/fixtures";
import {
    canSignLicenses,
    installLicense,
    issueKey,
    licensePayload,
    noSigner,
    readLicense,
    restoreColony,
    signKey,
    usageOf,
} from "../../../support/license";

// These change the installation's license, which every other test counts on
// being the Colony key env/up.sh installs: they run alone, after every other
// test (playwright.config.ts), one at a time, and put the Colony key back once
// they end.
test.describe.configure({ mode: "serial", timeout: 300_000 });
test.skip(!canSignLicenses, noSigner);

const HOUR = 3_600_000;
const DAY = 24 * HOUR;
const LICENSE_PAGE = "/system/license/overview/";

let installation = "";

test.beforeAll(async ({ api }) => {
    installation = (await readLicense(api)).installationId;
    expect(installation, "the installation's ID, which a key names").not.toBe("");
    // Whatever a run stopped halfway left, the first test replaces Colony.
    await restoreColony(api);
});

test.afterAll(async ({ api }) => {
    await restoreColony(api);
});

// planRegion is System › License's Plan section.
function planRegion(page: Page) {
    return page.getByRole("region", { name: "Plan", exact: true });
}

// banner is the license banner of that title, above every page.
function banner(page: Page, title: string | RegExp) {
    return page.getByRole("status", { name: title, exact: typeof title === "string" ? true : undefined });
}

// noErrorToast checks that a refusal was said by the dialog alone.
async function noErrorToast(page: Page): Promise<void> {
    await expect(page.locator('[data-sonner-toast][data-type="error"]')).toHaveCount(0);
}

// 2.133
test("a Hive trial key pasted on System › License is in force at once; its banner, dismissed, stays hidden", async ({
    page,
    api,
}) => {
    const key = issueKey({ plan: "hive", installation, trial: true, days: 3, licenseId: "lic_e2e_trial" });

    await page.goto(LICENSE_PAGE);
    const form = page.getByRole("region", { name: "Install a key" });
    await form.getByRole("textbox", { name: "License key" }).fill(key);
    await form.getByRole("button", { name: "Install key" }).click();
    const confirm = page.getByRole("alertdialog", { name: "Replace the license key?" });
    await expect(confirm).toContainText("This key replaces the Colony key in force, whatever its plan.");
    await confirm.getByRole("button", { name: "Replace key" }).click();

    await expect(page.locator('[data-sonner-toast][data-type="success"]')).toContainText(
        "License installed: Hive trial, until",
    );
    const plan = planRegion(page);
    await expect(plan.getByText("Hive", { exact: true })).toBeVisible();
    await expect(plan.getByText("Trial", { exact: true })).toBeVisible();
    await expect(plan).toContainText("A trial of Hive until");
    await expect(plan).toContainText("(3 days left). Then this installation is on Bee.");
    await expect(form.getByRole("textbox", { name: "License key" })).toHaveValue("");
    const license = await readLicense(api);
    expect(`${license.plan} ${license.status}`).toBe("hive trial");

    // Not on the License page, which says it already; on every other.
    const trial = banner(page, "Hive trial: 3 days left.");
    await expect(trial).toHaveCount(0);
    await page.goto("/home/");
    await expect(trial).toBeVisible();
    await expect(trial).toContainText("Then this installation is on Bee, the free plan.");
    await expect(trial.getByRole("link", { name: "Upgrade" })).toHaveAttribute(
        "href",
        license.portal?.pricingUrl ?? "no pricing page",
    );
    await expect(trial.getByRole("link", { name: "License", exact: true })).toHaveAttribute("href", LICENSE_PAGE);

    await trial.getByRole("button", { name: "Dismiss" }).click();
    await expect(trial).toHaveCount(0);

    // Kept in the browser: the page loaded again reads the license, and the
    // banner stays hidden.
    const read = page.waitForResponse(res => new URL(res.url()).pathname === "/api/system/license" && res.ok());
    await page.reload();
    await read;
    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    await page.waitForLoadState("networkidle");
    await expect(trial).toHaveCount(0);
});

// 2.134
test("a key past its paid period holds through its grace: the banner counts the days left", async ({ page, api }) => {
    const now = Date.now();
    // 28 days and 23 hours of grace left: 29, rounded up.
    const license = await installLicense(
        api,
        signKey(
            licensePayload({
                plan: "comb",
                installation,
                issuedAt: new Date(now - 31 * DAY),
                expiresAt: new Date(now - DAY - HOUR),
                graceDays: 30,
                licenseId: "lic_e2e_grace",
            }),
        ),
    );
    expect(`${license.plan} ${license.status}`).toBe("comb grace");

    await page.goto("/home/");
    const grace = banner(page, /^The Comb license ran out on .+: 29 days of grace left\.$/);
    await expect(grace).toBeVisible();
    await expect(grace).toContainText("Renew it, or paste the renewed key.");
    await expect(grace.getByRole("link", { name: "Renew" })).toHaveAttribute(
        "href",
        license.portal?.accountUrl ?? "no account page",
    );

    await grace.getByRole("link", { name: "License", exact: true }).click();
    await expect(page).toHaveURL(new RegExp(`${LICENSE_PAGE}$`));
    const plan = planRegion(page);
    await expect(plan.getByText("Comb", { exact: true })).toBeVisible();
    await expect(plan.getByText("Grace period", { exact: true })).toBeVisible();
    await expect(plan).toContainText("The paid period ended on");
    await expect(plan).toContainText("(29 days left), then this installation is on Bee.");
    await expect(plan.getByRole("group", { name: "Grace ends" })).toBeVisible();
});

// 2.135
test("over a limit, the banner says so; a join command and a new app are refused by the plan's dialog", async ({
    page,
    api,
    cleanup,
}) => {
    // More apps than the key will allow, whatever else the installation has.
    const project = await createProject(api, e2eName("license-limits"));
    cleanup(() => deleteProject(api, project.id));
    await createApp(api, project, "one");
    await createApp(api, project, "two");

    const now = Date.now();
    const license = await installLicense(
        api,
        signKey(
            licensePayload({
                plan: "comb",
                installation,
                issuedAt: new Date(now - HOUR),
                expiresAt: new Date(now + 30 * DAY),
                graceDays: 14,
                limits: { nodes: 1, apps: 1 },
                licenseId: "lic_e2e_limits",
            }),
        ),
    );
    expect(`${license.plan} ${license.status}`).toBe("comb valid");
    const apps = usageOf(license, "apps");
    expect(apps.used, "apps beyond the key's one").toBeGreaterThan(1);
    expect(usageOf(license, "nodes").used, "dind's one node").toBe(1);
    const pricingUrl = license.portal?.pricingUrl ?? "no pricing page";

    // Only apps are over: one node of one is at the limit, not past it.
    await page.goto("/home/");
    const limits = banner(page, `${apps.used} of 1 app: no app can be added.`);
    await expect(limits).toBeVisible();
    await expect(limits).toContainText("Nothing is removed. Remove some, or upgrade.");

    await test.step("Join Node: no join command past the plan's nodes", async () => {
        await page.goto("/cluster/nodes/");
        await page.getByRole("button", { name: "Join Node" }).click();
        await page.getByRole("button", { name: "Get Join Command" }).click();
        const refused = page.getByRole("dialog", { name: "No more nodes on Comb" });
        await expect(refused).toContainText(
            "The Comb plan allows 1 node, and the swarm has as many, so HivePaaS gives no join command.",
        );
        await expect(refused).toContainText("Now 1 of 1 node on Comb");
        await expect(refused.getByRole("link", { name: "Upgrade" })).toHaveAttribute("href", pricingUrl);
        await noErrorToast(page);
        await refused.getByRole("button", { name: "Close" }).first().click();
        await expect(refused).toHaveCount(0);
    });

    await test.step("New App: refused, and View license shows the usage", async () => {
        await page.goto(`/projects/${project.id}/apps/`);
        await page.getByRole("button", { name: "New App" }).click();
        const create = page.getByRole("dialog", { name: "Create App" });
        await create.getByRole("group", { name: "Name *" }).getByRole("textbox").fill("three");
        await create.getByRole("button", { name: "Create App" }).click();

        const refused = page.getByRole("dialog", { name: "No more apps on Comb" });
        await expect(refused).toContainText("The Comb plan allows 1 app.");
        await expect(refused).toContainText(`Now ${apps.used} of 1 app on Comb`);
        await noErrorToast(page);
        const near = { id: "", name: "", projectId: project.id, env: "development" };
        expect(await findAppNamed(api, near, "three"), "no app made").toBeUndefined();

        await refused.getByRole("link", { name: "View license" }).click();
        await expect(page).toHaveURL(new RegExp(`${LICENSE_PAGE}$`));
        await expect(refused).toHaveCount(0);
        await expect(create).toHaveCount(0);
        const usage = page.getByRole("region", { name: "Usage" });
        await expect(usage.getByRole("meter", { name: "Apps" })).toHaveAttribute(
            "aria-valuetext",
            `${apps.used} of 1 app`,
        );
        await expect(usage.getByRole("group", { name: "Apps", exact: true })).toContainText(
            "Over the limit: none can be added",
        );
        await expect(usage.getByRole("group", { name: "Nodes", exact: true })).toContainText("Full");
        // The limits banner stays on the License page, without its link to it.
        await expect(limits).toBeVisible();
        await expect(limits.getByRole("link", { name: "License", exact: true })).toHaveCount(0);
    });
});

// createGoogleSignIn makes an active Google sign-in provider, which only a paid
// plan allows, and answers its id.
async function createGoogleSignIn(api: APIRequestContext, name: string): Promise<string> {
    const body = (await ok(
        await api.post("settings/oauth", {
            data: {
                kind: "google",
                name,
                clientId: "e2e-client-id",
                clientSecret: "e2e-client-secret-0123456789",
                scopes: [],
            },
        }),
        `creating the Google sign-in ${name}`,
    )) as { data: { id: string } };
    return body.data.id;
}

// disableSignIn disables a sign-in provider, which Bee allows.
async function disableSignIn(api: APIRequestContext, id: string): Promise<void> {
    const current = (await ok(await api.get(`settings/oauth/${id}`), "reading the sign-in provider")) as {
        data: { updateVer: number };
    };
    await ok(
        await api.put(`settings/oauth/${id}/status`, {
            data: { status: "disabled", updateVer: current.data.updateVer },
        }),
        "disabling the sign-in provider",
    );
}

// 2.136
test("once a key runs out the installation is on Bee: a Google sign-in set up before is kept, none can be added or enabled", async ({
    page,
    api,
    cleanup,
}) => {
    const now = Date.now();
    // A trial has no grace: it runs out on its date, in 45 seconds.
    const trial = await installLicense(
        api,
        signKey(
            licensePayload({
                plan: "hive",
                installation,
                issuedAt: new Date(now - HOUR),
                expiresAt: new Date(now + 45_000),
                graceDays: 0,
                trial: true,
                licenseId: "lic_e2e_runs_out",
            }),
        ),
    );
    expect(`${trial.plan} ${trial.status}`).toBe("hive trial");

    // Set up while the trial lets it.
    const name = e2eName("google");
    cleanup(() => deleteSettingsNamed(api, "oauth", name));
    const googleId = await createGoogleSignIn(api, name);

    await expect
        .poll(async () => (await readLicense(api)).status, { timeout: 90_000, intervals: [5_000] })
        .toBe("expired");
    const license = await readLicense(api);
    expect(license.plan).toBe("bee");

    await test.step("the banner, and the License page, say it is on Bee", async () => {
        await page.goto("/home/");
        const expired = banner(page, "The license has expired: this installation is on Bee.");
        await expect(expired).toBeVisible();
        await expect(expired).toContainText("Everything keeps running; nothing more can be added past Bee's limits.");
        await expect(expired.getByRole("link", { name: "Upgrade" })).toHaveAttribute(
            "href",
            license.portal?.pricingUrl ?? "no pricing page",
        );
        // Bee allows 2 users, fewer than the installation has.
        await expect(banner(page, /\d+ of 2 users: no user can be invited or re-enabled/)).toBeVisible();

        await expired.getByRole("link", { name: "License", exact: true }).click();
        const plan = planRegion(page);
        await expect(plan.getByText("Bee", { exact: true })).toBeVisible();
        await expect(plan.getByText("Expired", { exact: true })).toBeVisible();
        await expect(plan).toContainText("The key ran out on");
        await expect(plan.getByRole("group", { name: "Key plan" })).toContainText("Hive trial");
        // A trial is offered only to an installation that never had a key.
        await expect(plan.getByRole("link", { name: "Start a 14-day trial" })).toHaveCount(0);
        await expect(
            page.getByRole("region", { name: "Usage" }).getByRole("group", {
                name: "Google, Microsoft and OpenID Connect sign-in",
            }),
        ).toContainText(/Not included on Bee · \d+ already set up keeps? signing people in/);
    });

    await test.step("Settings › OAuth: the Google sign-in is kept, and no other can be added", async () => {
        await page.goto("/integrations/oauth/");
        await expect(page.getByRole("row").filter({ hasText: name })).toContainText("Kept on Bee");

        await page.goto("/integrations/oauth/create/");
        await page.getByRole("group", { name: "Provider" }).getByRole("combobox").click();
        await expect(page.getByRole("option", { name: "Google · paid plan" })).toHaveAttribute("aria-disabled", "true");
        await expect(page.getByRole("option", { name: "Github", exact: true })).not.toHaveAttribute(
            "aria-disabled",
            "true",
        );
        await page.keyboard.press("Escape");
        await expect(
            page.getByText(
                "Google, Microsoft Online and OpenID Connect sign-in need a paid plan: this installation is on Bee.",
            ),
        ).toBeVisible();
    });

    await test.step("disabled, the Google sign-in cannot be enabled again", async () => {
        await disableSignIn(api, googleId);
        await page.goto("/integrations/oauth/");
        const row = page.getByRole("row").filter({ hasText: name });
        await expect(row).toContainText("Unavailable on Bee");
        await row.getByRole("button", { name: "Actions menu" }).click();
        await page.getByRole("button", { name: "Change Status" }).click();
        const dialog = page.getByRole("dialog", { name: "Change status" });
        await expect(dialog.getByRole("tab", { name: "Active" })).toBeDisabled();
        await expect(dialog.getByRole("tab", { name: "Disabled" })).toHaveAttribute("aria-selected", "true");
        await expect(dialog).toContainText("A paid plan is needed to enable it.");
    });
});
