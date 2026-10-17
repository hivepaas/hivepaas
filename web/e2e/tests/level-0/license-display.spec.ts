import type { Page } from "@playwright/test";

import { expect, test } from "../../support/fixtures";
import type { License, LicenseUsage } from "../../support/license";

// What the dashboard shows of a license the installation under test cannot be
// put in: no key at all (there is no way to remove one), a key moved to
// another installation by the portal, a key stored for another installation,
// one that no longer verifies - and the plan-limit dialog's words for each
// limit. The license, and the refusals, are answered by the test (page.route),
// never by the server: this alone of the suite answers for the API, and only
// for these. A trial, a grace period, a key that ran out, a limit passed and
// the refusals that follow are put in force for real, in
// tests/level-2/license-on/. Nothing is written.

const LICENSE_PAGE = "/system/license/overview/";
const HOUR = 3_600_000;
const DAY = 24 * HOUR;

// The docs' pages on plans and keys, which the dashboard leads to.
const LICENSE_DOCS = "https://docs.hivepaas.com/plans-and-licenses";

const INSTALLATION = "e2efixtureinstallation01";
const PORTAL = {
    url: "https://portal.e2e.example",
    pricingUrl: `https://portal.e2e.example/pricing?installationId=${INSTALLATION}`,
    trialUrl: `https://portal.e2e.example/trial?installationId=${INSTALLATION}`,
    accountUrl: "https://portal.e2e.example/account/licenses",
};

// at is a moment from now, as the API writes it.
function at(fromNow: number): string {
    return new Date(Date.now() + fromNow).toISOString();
}

// Bee's limits, none of them passed.
function beeLimits(): LicenseUsage[] {
    return [
        { limit: "nodes", max: 2, used: 1 },
        { limit: "users", max: 2, used: 2 },
        { limit: "apps", max: 20, used: 3 },
        { limit: "functions", max: 20, used: 0 },
        { limit: "previews", max: 2, used: 0 },
        { limit: "sign-in-providers", max: 0, used: 0 },
    ];
}

// Bee's limits, with as many users, functions and previews as it allows: what
// a refusal of one more of them finds.
function beeFull(): LicenseUsage[] {
    return [
        { limit: "nodes", max: 2, used: 1 },
        { limit: "users", max: 2, used: 2 },
        { limit: "apps", max: 20, used: 3 },
        { limit: "functions", max: 20, used: 20 },
        { limit: "previews", max: 2, used: 2 },
        { limit: "sign-in-providers", max: 0, used: 1 },
    ];
}

// A paid plan's limits: nodes and users counted, the rest not.
function paidLimits(nodes: number, users: number): LicenseUsage[] {
    return [
        { limit: "nodes", max: nodes, used: 1 },
        { limit: "users", max: users, used: 2 },
        { limit: "apps", max: null, used: 3 },
        { limit: "functions", max: null, used: 0 },
        { limit: "previews", max: null, used: 0 },
        { limit: "sign-in-providers", max: null, used: 0 },
    ];
}

// licenseOf is a license as the API answers it: by default a Colony key in
// force for another 20 days, with what is set over it.
function licenseOf(over: Partial<License>): License {
    return {
        status: "valid",
        plan: "colony",
        keyPlan: "colony",
        trial: false,
        development: false,
        licenseId: "lic_fixture",
        customerName: "Fixture Ltd",
        installationId: INSTALLATION,
        installations: [INSTALLATION],
        issuedAt: at(-10 * DAY),
        expiresAt: at(20 * DAY),
        graceEndsAt: at(34 * DAY),
        limits: paidLimits(25, 30),
        portal: PORTAL,
        refresh: { enabled: true },
        ...over,
    };
}

// answerLicense answers the dashboard's every read of the license with this
// one, and anything else asked of it with a refusal: nothing reaches the
// server's. It answers how many reads it served.
async function answerLicense(page: Page, license: License): Promise<() => number> {
    let served = 0;
    await page.route(
        url => url.pathname === "/api/system/license",
        route => {
            if (route.request().method() !== "GET") return route.abort();
            served++;
            return route.fulfill({ json: { meta: {}, data: license } });
        },
    );
    return () => served;
}

// The license banner of that title, above every page.
function banner(page: Page, title: string) {
    return page.getByRole("status", { name: title, exact: true });
}

interface StatusCase {
    what: string;
    license: () => License;
    // The status banner's title and text; none for no key.
    banner?: { title: string; body: string };
    plan: string;
    badge: string;
    // Parts of the status sentence.
    sentence: string[];
    // Rows of the key's details, by title, with what they hold ("": shown).
    rows?: [title: string, holds: string][];
}

const OTHER_INSTALLATION = "e2eotherinstallation0001";

const STATUSES: StatusCase[] = [
    {
        what: "no key",
        license: () =>
            licenseOf({
                status: "none",
                plan: "bee",
                keyPlan: "",
                licenseId: "",
                customerName: "",
                installations: [],
                issuedAt: undefined,
                expiresAt: undefined,
                graceEndsAt: undefined,
                limits: beeLimits(),
            }),
        plan: "Bee",
        badge: "No license",
        sentence: ["This installation runs on Bee, the free plan."],
    },
    {
        what: "a key moved to another installation, within the grace the move leaves",
        license: () =>
            licenseOf({
                status: "moved",
                movedAt: at(-2 * DAY),
                graceEndsAt: at(12 * DAY - HOUR),
            }),
        banner: {
            title: "This license key now belongs to another installation.",
            body: "It holds here 12 more days, then this installation is on Bee.",
        },
        plan: "Colony",
        badge: "Moved",
        sentence: [
            "The portal moved this key to another installation on",
            "(12 days left), then this installation is on Bee.",
        ],
        rows: [
            ["Moved", ""],
            ["Grace ends", ""],
        ],
    },
    {
        what: "a key moved to another installation, past the grace",
        license: () =>
            licenseOf({
                status: "moved",
                plan: "bee",
                movedAt: at(-20 * DAY),
                graceEndsAt: at(-6 * DAY),
                limits: beeLimits(),
            }),
        banner: {
            title: "This license key now belongs to another installation: this installation is on Bee.",
            body: "Everything keeps running; nothing more can be added past Bee's limits.",
        },
        plan: "Bee",
        badge: "Moved",
        sentence: ["The portal moved this key to another installation on", "This installation is now on Bee."],
        rows: [["Key plan", "Colony"]],
    },
    {
        what: "another installation's key",
        license: () =>
            licenseOf({
                status: "other-installation",
                plan: "bee",
                keyPlan: "hive",
                installations: [OTHER_INSTALLATION],
                limits: beeLimits(),
            }),
        banner: {
            title: "The stored license key is for another installation: this installation is on Bee.",
            body: "Everything keeps running; nothing more can be added past Bee's limits.",
        },
        plan: "Bee",
        badge: "Another installation's key",
        sentence: ["The stored key is for another installation: this one is on Bee."],
        rows: [
            ["Key is for", OTHER_INSTALLATION],
            ["Key plan", "Hive"],
        ],
    },
    {
        what: "a key that does not verify",
        license: () =>
            licenseOf({
                status: "invalid",
                plan: "bee",
                keyPlan: "",
                licenseId: "",
                customerName: "",
                installations: [],
                issuedAt: undefined,
                expiresAt: undefined,
                graceEndsAt: undefined,
                reason: "the ml-dsa-65 signature does not verify",
                limits: beeLimits(),
            }),
        banner: {
            title: "The stored license key cannot be verified: this installation is on Bee.",
            body: "Everything keeps running; nothing more can be added past Bee's limits.",
        },
        plan: "Bee",
        badge: "Invalid key",
        sentence: ["The stored key could not be verified: this installation is on Bee."],
        rows: [["Reason", "the ml-dsa-65 signature does not verify"]],
    },
];

for (const c of STATUSES) {
    test(`${c.what}: the banner and System › License say what it means`, async ({ page }) => {
        const served = await answerLicense(page, c.license());

        await page.goto("/home/");
        await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
        await expect.poll(served, { message: "the license read" }).toBeGreaterThan(0);
        if (c.banner) {
            const shown = banner(page, c.banner.title);
            await expect(shown).toBeVisible();
            await expect(shown).toContainText(c.banner.body);
            await expect(shown.getByRole("link", { name: "License", exact: true })).toHaveAttribute(
                "href",
                LICENSE_PAGE,
            );
        } else {
            await page.waitForLoadState("networkidle");
            await expect(page.getByRole("button", { name: "Dismiss" })).toHaveCount(0);
        }

        await page.goto(LICENSE_PAGE);
        const plan = page.getByRole("region", { name: "Plan", exact: true });
        await expect(plan.getByText(c.plan, { exact: true })).toBeVisible();
        // The badge, first: a row of the key's details can bear the same word.
        await expect(plan.getByText(c.badge, { exact: true }).first()).toBeVisible();
        for (const part of c.sentence) {
            await expect(plan).toContainText(part);
        }
        for (const [title, holds] of c.rows ?? []) {
            const row = plan.getByRole("group", { name: title });
            await (holds === "" ? expect(row).toBeVisible() : expect(row).toContainText(holds));
        }
        // The status banner is left off the page that says the same.
        if (c.banner) {
            await expect(banner(page, c.banner.title)).toHaveCount(0);
        }
    });
}

test("no key: a trial is offered, and there is nothing to check for a newer key", async ({ page }) => {
    await answerLicense(page, STATUSES[0]!.license());

    await page.goto(LICENSE_PAGE);
    const plan = page.getByRole("region", { name: "Plan", exact: true });
    const trial = plan.getByRole("link", { name: "Start a 14-day trial" });
    await expect(trial).toHaveAttribute("href", PORTAL.trialUrl);
    await expect(trial).toHaveAttribute("target", "_blank");
    await expect(plan.getByRole("link", { name: "Upgrade" })).toHaveAttribute("href", PORTAL.pricingUrl);
    await expect(plan.getByRole("link", { name: "Your licenses" })).toHaveCount(0);
    await expect(page.getByRole("button", { name: "Check for a newer key" })).toBeDisabled();

    const usage = page.getByRole("region", { name: "Usage" });
    await expect(usage.getByRole("meter", { name: "Nodes" })).toHaveAttribute("aria-valuetext", "1 of 2 nodes");
    await expect(usage.getByRole("group", { name: "Users", exact: true })).toContainText("Full");
    await expect(usage.getByRole("group", { name: "Google, Microsoft and OpenID Connect sign-in" })).toContainText(
        "Not included on Bee",
    );
});

// Each docs page where it is wanted; and, beside the daily check, where it is
// turned off - hivepaas.toml, which outlasts a redeploy, or the variable.
test("the License page leads to the docs, and says where the daily check is turned off", async ({ page }) => {
    await answerLicense(page, licenseOf({ refresh: { enabled: false } }));

    await page.goto(LICENSE_PAGE);
    const plan = page.getByRole("region", { name: "Plan", exact: true });
    await expect(plan.getByRole("link", { name: "Plans and limits" })).toHaveAttribute(
        "href",
        `${LICENSE_DOCS}/plans-and-limits`,
    );
    const install = page.getByRole("region", { name: "Install a key" });
    await expect(install.getByRole("link", { name: "Installing a license key" })).toHaveAttribute(
        "href",
        `${LICENSE_DOCS}/install-a-license-key`,
    );

    const renewal = page.getByRole("region", { name: "Online renewal" });
    const daily = renewal.getByRole("group", { name: "Daily check" });
    await expect(daily).toContainText("Off");
    await expect(daily).toContainText("Turned off with refresh_disabled = true under [license] in hivepaas.toml");
    await expect(daily).toContainText("or with HP_LICENSE_REFRESH_DISABLED=true");
    const sent = daily.getByRole("link", { name: "What is sent, and turning it off" });
    await expect(sent).toHaveAttribute("href", `${LICENSE_DOCS}/renewal-and-privacy`);
    await expect(sent).toHaveAttribute("target", "_blank");
});

test("over several limits: one banner says what can no longer be done for each", async ({ page }) => {
    await answerLicense(
        page,
        licenseOf({
            plan: "comb",
            keyPlan: "comb",
            limits: [
                { limit: "nodes", max: 5, used: 7 },
                { limit: "users", max: 5, used: 6 },
                { limit: "apps", max: null, used: 3 },
                { limit: "functions", max: null, used: 0 },
                { limit: "previews", max: 1, used: 2 },
                { limit: "sign-in-providers", max: null, used: 0 },
            ],
        }),
    );

    await page.goto("/home/");
    const over = banner(
        page,
        "7 of 5 nodes: no node can be added; 6 of 5 users: no user can be invited or re-enabled; " +
            "2 of 1 preview: no preview can be opened.",
    );
    await expect(over).toBeVisible();
    await expect(over).toContainText("Nothing is removed. Remove some, or upgrade.");
    await expect(over.getByRole("link", { name: "Upgrade" })).toHaveAttribute("href", PORTAL.pricingUrl);

    await page.goto(LICENSE_PAGE);
    await expect(page.getByRole("region", { name: "Usage" }).getByRole("meter", { name: "Nodes" })).toHaveAttribute(
        "aria-valuetext",
        "7 of 5 nodes",
    );
});

interface RefusalCase {
    what: string;
    // The license the dialog reads as it opens.
    license: () => License;
    // The refusal's params, as the server exposes them; none at all for one.
    params?: Record<string, unknown>;
    detail: string;
    title: string;
    body: string;
    // The usage the dialog reads from the license, when it has one to show.
    usage?: string;
}

// The installation on Bee, with no key, as full as it may be.
function onBee(): License {
    return licenseOf({
        status: "none",
        plan: "bee",
        keyPlan: "",
        licenseId: "",
        customerName: "",
        installations: [],
        issuedAt: undefined,
        expiresAt: undefined,
        graceEndsAt: undefined,
        limits: beeFull(),
    });
}

// A Comb key that allows no functions.
function onCombWithoutFunctions(): License {
    return licenseOf({
        plan: "comb",
        keyPlan: "comb",
        limits: paidLimits(5, 5).map(u => (u.limit === "functions" ? { ...u, max: 0 } : u)),
    });
}

// The refusals 2.135 does not meet for real: it passes the nodes and apps
// limits.
const REFUSALS: RefusalCase[] = [
    {
        what: "users",
        license: onBee,
        params: { Limit: "users", Max: 2, Plan: "bee" },
        detail: "The Bee plan allows 2 users. Remove one, or upgrade.",
        title: "No more users on Bee",
        body: "The Bee plan allows 2 users, counting those invited who have not signed up yet. Disable a user or withdraw an invitation to free a seat.",
        usage: "Now 2 of 2 users on Bee",
    },
    {
        what: "functions",
        license: onBee,
        params: { Limit: "functions", Max: 20, Plan: "bee" },
        detail: "The Bee plan allows 20 functions. Remove one, or upgrade.",
        title: "No more functions on Bee",
        body: "The Bee plan allows 20 functions. Delete one to make room.",
        usage: "Now 20 of 20 functions on Bee",
    },
    {
        what: "previews",
        license: onBee,
        params: { Limit: "previews", Max: 2, Plan: "bee" },
        detail: "The Bee plan allows 2 previews at a time. Remove one, or upgrade.",
        title: "No more previews at a time on Bee",
        body: "The Bee plan runs 2 previews at a time. Close one to open another.",
        usage: "Now 2 of 2 previews on Bee",
    },
    {
        what: "paid sign-in on Bee",
        license: onBee,
        params: { Limit: "sign-in-providers", Max: 0, Plan: "bee" },
        detail: "The Bee plan does not include Google, Microsoft or OpenID Connect sign-in providers. Upgrade to add one.",
        title: "Google, Microsoft and OpenID Connect sign-in need a paid plan",
        body: "Bee does not include them. The ones already set up keep signing people in; none can be added or enabled.",
    },
    {
        what: "a kind the plan leaves out",
        license: onCombWithoutFunctions,
        params: { Limit: "functions", Max: 0, Plan: "comb" },
        detail: "The Comb plan does not include functions. Upgrade to add one.",
        title: "No more functions on Comb",
        body: "The Comb plan does not include functions.",
    },
    {
        what: "a kind this dashboard does not know",
        license: onCombWithoutFunctions,
        params: { Limit: "gpus", Max: 1, Plan: "comb" },
        detail: "The Comb plan allows 1 GPU. Remove one, or upgrade.",
        title: "Plan limit reached",
        body: "The Comb plan allows 1 GPU. Remove one, or upgrade.",
    },
    {
        what: "no params",
        license: onBee,
        detail: "The plan in force does not allow it.",
        title: "Plan limit reached",
        body: "The plan in force does not allow it.",
    },
];

// refusal is a 412 as the server answers a creation the plan does not allow.
function refusal(c: RefusalCase) {
    return {
        status: 412,
        json: {
            title: "Precondition Failed",
            status: 412,
            code: "ERR_PLAN_LIMIT_REACHED",
            detail: c.detail,
            ...(c.params ? { params: c.params } : {}),
        },
    };
}

// The words come from the refusal alone, wherever it was met: Join Node's
// command, a read that writes nothing, stands for every creation.
for (const c of REFUSALS) {
    test(`a creation refused for ${c.what}: the plan-limit dialog says why, and leads to the plan`, async ({
        page,
    }) => {
        await answerLicense(page, c.license());
        await page.route(
            url => url.pathname === "/api/cluster/nodes/join-command",
            route => route.fulfill(refusal(c)),
        );

        await page.goto("/cluster/nodes/");
        await page.getByRole("button", { name: "Join Node" }).click();
        await page.getByRole("button", { name: "Get Join Command" }).click();

        const dialog = page.getByRole("dialog", { name: c.title });
        await expect(dialog).toContainText(c.body);
        if (c.usage) {
            await expect(dialog).toContainText(c.usage);
        } else {
            await expect(dialog.getByText(/^Now \d+ of/)).toHaveCount(0);
        }
        await expect(dialog.getByRole("link", { name: "Upgrade" })).toHaveAttribute("href", PORTAL.pricingUrl);
        await expect(dialog.getByRole("link", { name: "View license" })).toHaveAttribute("href", LICENSE_PAGE);
        await expect(dialog.getByRole("link", { name: "Plans and limits" })).toHaveAttribute(
            "href",
            `${LICENSE_DOCS}/plans-and-limits`,
        );
        await expect(page.locator('[data-sonner-toast][data-type="error"]')).toHaveCount(0);

        await dialog.getByRole("button", { name: "Close" }).first().click();
        await expect(dialog).toHaveCount(0);
    });
}

// A Google sign-in saved while the page held the plan to include it - a key
// that ran out since, say - is refused by the server; the form's own failure
// opens the dialog, not a toast.
test("a Google sign-in the server refuses on Bee opens the plan-limit dialog from the OAuth form", async ({ page }) => {
    await answerLicense(page, licenseOf({ plan: "hive", keyPlan: "hive", limits: paidLimits(10, 10) }));
    let refused = 0;
    await page.route(
        url => url.pathname === "/api/settings/oauth",
        route => {
            if (route.request().method() !== "POST") return route.fallback();
            refused++;
            return route.fulfill(refusal(REFUSALS.find(r => r.what === "paid sign-in on Bee")!));
        },
    );

    await page.goto("/integrations/oauth/create/");
    const field = (label: RegExp) => page.getByRole("group", { name: label });
    await field(/^Provider/)
        .getByRole("combobox")
        .click();
    await page.getByRole("option", { name: "Google", exact: true }).click();
    await field(/^Name/).getByRole("textbox").fill("e2e-google-refused");
    await field(/^Organization/)
        .getByRole("textbox")
        .fill("e2e");
    await field(/^Client ID/)
        .getByRole("textbox")
        .fill("e2e-client-id");
    await field(/^Client Secret/)
        .locator("input")
        .fill("e2e-client-secret");
    await page.getByRole("button", { name: "Save" }).click();

    const dialog = page.getByRole("dialog", { name: "Google, Microsoft and OpenID Connect sign-in need a paid plan" });
    await expect(dialog).toContainText("Bee does not include them.");
    await expect(dialog.getByRole("link", { name: "Upgrade" })).toHaveAttribute("href", PORTAL.pricingUrl);
    expect(refused, "the one save, refused").toBe(1);
    await expect(page.locator('[data-sonner-toast][data-type="error"]')).toHaveCount(0);
});
