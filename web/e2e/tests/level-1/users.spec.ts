import {
    appPath,
    createApp,
    createProject,
    deleteProject,
    deleteUsersByEmail,
    findApp,
    runtimeEnvVars,
    setRuntimeEnvVars,
} from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";
import { signIn } from "../../support/sign-in";
import { totp, wrongTotp } from "../../support/totp";
import { PASSWORD, SIGNED_OUT, inviteAndSignUp, signedInAs } from "../../support/users";

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
