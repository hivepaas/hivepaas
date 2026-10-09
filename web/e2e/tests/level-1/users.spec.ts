import { type APIRequestContext, type Browser, type Page, request } from "@playwright/test";

import {
    type Project,
    appPath,
    createApp,
    createProject,
    deleteProject,
    deleteUsersByEmail,
    findApp,
    runtimeEnvVars,
    setRuntimeEnvVars,
} from "../../support/api";
import { env } from "../../support/env";
import { e2eName, expect, test } from "../../support/fixtures";
import { signIn } from "../../support/sign-in";
import { totp, wrongTotp } from "../../support/totp";

const SIGNED_OUT = { baseURL: env.baseURL, storageState: { cookies: [], origins: [] } };
const PASSWORD = "E2e-Passw0rd-member";

interface Member {
    username: string;
    email: string;
}

// inviteAndSignUp has an admin invite a member to one project, by a link
// rather than an email, and the member sign up from it.
async function inviteAndSignUp(page: Page, browser: Browser, label: string, project: Project): Promise<Member> {
    const member = { username: e2eName(label).replace(/-/g, "_"), email: `${e2eName(label)}@example.com` };

    const link = await test.step("an admin invites a member to one project", async () => {
        await page.goto("/user-management/users/");
        await page.getByRole("button", { name: "Invite User" }).click();
        const dialog = page.getByRole("dialog");
        await dialog.getByRole("group", { name: "Email" }).getByRole("textbox").fill(member.email);
        await dialog.getByRole("group", { name: "Project Access" }).getByRole("combobox").click();
        await page.getByRole("option", { name: project.name }).click();
        await page.keyboard.press("Escape");
        await dialog.getByRole("group", { name: "Project Access" }).getByRole("button", { name: "Add" }).click();
        await dialog.getByRole("button", { name: "Generate Invite Link" }).click();
        return dialog.getByText(/\/auth\/sign-up\/?\?token=/).innerText();
    });

    await test.step("the member signs up from the link", async () => {
        const signUp = await browser.newPage(SIGNED_OUT);
        try {
            await signUp.goto(link);
            await signUp.getByRole("textbox", { name: "Username *" }).fill(member.username);
            await signUp.getByRole("textbox", { name: "Full Name *" }).fill("E2E Member");
            await signUp.getByRole("textbox", { name: "Password *", exact: true }).fill(PASSWORD);
            await signUp.getByRole("textbox", { name: "Retype Password *" }).fill(PASSWORD);
            await signUp.getByRole("checkbox", { name: "Agree to Terms and Conditions" }).check();
            await signUp.getByRole("button", { name: "Sign up" }).click();
            // Signed up, then asked to sign in.
            await expect(signUp).toHaveURL(/\/auth\/sign-in\//);
        } finally {
            await signUp.close();
        }
    });
    return member;
}

test("an invited member signs up, and sees only the project given", async ({ page, api, cleanup, browser }) => {
    const granted = await createProject(api, e2eName("granted"));
    cleanup(() => deleteProject(api, granted.id));
    const withheld = await createProject(api, e2eName("withheld"));
    cleanup(() => deleteProject(api, withheld.id));
    cleanup(() => deleteUsersByEmail(api, `${e2eName("member")}@example.com`));

    const member = await inviteAndSignUp(page, browser, "member", granted);

    const asMember = await browser.newPage(SIGNED_OUT);
    try {
        await asMember.goto("/");
        await signIn(asMember, member.username, PASSWORD);
        await expect(asMember).toHaveURL(/\/home\/$/);
        await asMember.goto("/projects/");
        await expect(asMember.getByRole("row").filter({ hasText: granted.name })).toBeVisible();
        await expect(asMember.getByRole("row").filter({ hasText: withheld.name })).toHaveCount(0);
    } finally {
        await asMember.close();
    }
});

test("a disabled member is turned away at sign-in", async ({ page, api, cleanup, browser }) => {
    const project = await createProject(api, e2eName("disabled-project"));
    cleanup(() => deleteProject(api, project.id));
    cleanup(() => deleteUsersByEmail(api, `${e2eName("disabled")}@example.com`));
    const member = await inviteAndSignUp(page, browser, "disabled", project);

    await page.goto("/user-management/users/");
    await page.getByRole("row").filter({ hasText: member.email }).getByRole("button", { name: "User menu" }).click();
    await page.getByRole("menu").getByRole("button", { name: "Disable User" }).click();
    await page
        .getByRole("dialog")
        .getByRole("button", { name: /Disable/ })
        .click();
    await expect(page.getByRole("row").filter({ hasText: member.email })).toContainText("Disabled");

    const again = await browser.newPage(SIGNED_OUT);
    try {
        await again.goto("/");
        const login = again.waitForResponse(res => res.url().includes("/api/auth/login-with-password"));
        await signIn(again, member.username, PASSWORD);

        expect((await login).ok(), "signing in as a disabled user is refused").toBe(false);
        await expect(again.getByText("User is unavailable").first()).toBeVisible();
        await expect(again).toHaveURL(/\/auth\/sign-in\//);
    } finally {
        await again.close();
    }
});

// signedInAs is the API as the user signs in to it.
async function signedInAs(username: string, password: string): Promise<APIRequestContext> {
    const anonymous = await request.newContext({ baseURL: `${env.baseURL}/api/` });
    const res = await anonymous.post("auth/login-with-password", { data: { username, password } });
    expect(res.ok(), `signing in as ${username}: ${res.status()}`).toBe(true);
    const token = ((await res.json()) as { data: { session: { accessToken: string } } }).data.session.accessToken;
    await anonymous.dispose();
    return request.newContext({
        baseURL: `${env.baseURL}/api/`,
        extraHTTPHeaders: { Authorization: `Bearer ${token}` },
    });
}

// A project given in the invitation is given to read: the member sees its apps
// and their settings, and changes nothing - Save is out of reach on the page,
// and the API refuses what the page would send, as it would a script's.
test("a member who may only read a project sees its apps, and is refused every change", async ({
    page,
    api,
    cleanup,
    browser,
}) => {
    const project = await createProject(api, e2eName("read-only"));
    cleanup(() => deleteProject(api, project.id));
    const app = await createApp(api, project, "web");
    const greeting = { key: "GREETING", value: "hello", isLiteral: true };
    await setRuntimeEnvVars(api, app, [greeting]);
    cleanup(() => deleteUsersByEmail(api, `${e2eName("reader")}@example.com`));
    const member = await inviteAndSignUp(page, browser, "reader", project);

    const asMember = await browser.newPage(SIGNED_OUT);
    try {
        await asMember.goto("/");
        await signIn(asMember, member.username, PASSWORD);
        await expect(asMember).toHaveURL(/\/home\/$/);
        await asMember.goto(`/${appPath(app)}/env-variables/`);
        const runtime = asMember.getByRole("region", { name: "Runtime Env Variables", exact: true });
        await expect(runtime.getByRole("group", { name: "GREETING", exact: true })).toBeVisible();
        await expect(asMember.getByRole("button", { name: "Save" })).toBeDisabled();
    } finally {
        await asMember.close();
    }

    const memberApi = await signedInAs(member.username, PASSWORD);
    try {
        expect((await memberApi.get(`${appPath(app)}/env-vars`)).status(), "the member reads").toBe(200);
        const refusals = {
            "saving its variables": await memberApi.put(`${appPath(app)}/env-vars`, {
                data: { updateVer: 1, runtimeEnvVars: [], buildtimeEnvVars: [], sharedEnvVars: [] },
            }),
            "deleting it": await memberApi.delete(appPath(app)),
            "making another": await memberApi.post(`projects/${project.id}/${app.env}/apps`, {
                data: { name: "another", env: app.env, note: "", tags: [], status: "active" },
            }),
        };
        for (const [what, res] of Object.entries(refusals)) {
            expect(res.status(), `the member is refused ${what}`).toBe(401);
            expect(((await res.json()) as { code: string }).code).toBe("ERR_UNAUTHORIZED");
        }
    } finally {
        await memberApi.dispose();
    }
    expect(await runtimeEnvVars(api, app), "its variables are as they were").toEqual([
        expect.objectContaining(greeting),
    ]);
    expect(await findApp(api, app), "the app is still there").toBeDefined();
});

// A member turns two-factor sign-in on from their profile, with the code their
// authenticator shows - computed here from the secret, as the app does. From
// then on the password is not enough: a code is asked for, a wrong one is
// refused - and tried again - and the right one lets them in.
test("a member who turns on two-factor sign-in is asked for a code, and only the right one lets them in", async ({
    page,
    api,
    cleanup,
    browser,
}) => {
    test.setTimeout(60_000);
    const project = await createProject(api, e2eName("two-factor"));
    cleanup(() => deleteProject(api, project.id));
    cleanup(() => deleteUsersByEmail(api, `${e2eName("two-factor")}@example.com`));
    const member = await inviteAndSignUp(page, browser, "two-factor", project);

    const setUp = await browser.newPage(SIGNED_OUT);
    let secret = "";
    try {
        await setUp.goto("/");
        await signIn(setUp, member.username, PASSWORD);
        await expect(setUp).toHaveURL(/\/home\/$/);
        await setUp.goto("/current-user/profile/");
        await setUp.getByRole("button", { name: "Activate 2FA" }).click();
        const dialog = setUp.getByRole("dialog", { name: "Activate 2FA" });
        await dialog.getByRole("button", { name: "Show code" }).click();
        secret = (await dialog.locator("p.select-all").innerText()).trim();
        expect(secret, "the secret, in base32").toMatch(/^[A-Z2-7]+=*$/);
        await dialog.getByLabel("Then, enter the passcode here").fill(totp(secret));
        await dialog.getByRole("button", { name: "Activate" }).click();
        await expect(setUp.getByText("2FA setup completed successfully")).toBeVisible();
    } finally {
        await setUp.close();
    }

    const again = await browser.newPage(SIGNED_OUT);
    try {
        await again.goto("/");
        await signIn(again, member.username, PASSWORD);
        const code = again.getByLabel("Authentication code");
        await expect(code, "a code is asked for").toBeVisible();

        await code.fill(wrongTotp(secret));
        await again.getByRole("button", { name: "Verify Code" }).click();
        await expect(again.getByText("Your Authentication Failed")).toBeVisible();
        await expect(again).not.toHaveURL(/\/home\/$/);

        await again.getByRole("button", { name: "Try Again" }).click();
        await code.fill(totp(secret));
        await again.getByRole("button", { name: "Verify Code" }).click();
        await expect(again).toHaveURL(/\/home\/$/);
    } finally {
        await again.close();
    }
});
