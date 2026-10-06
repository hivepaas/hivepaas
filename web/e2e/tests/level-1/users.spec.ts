import type { Browser, Page } from "@playwright/test";

import { type Project, createProject, deleteProject, deleteUsersByEmail } from "../../support/api";
import { env } from "../../support/env";
import { e2eName, expect, test } from "../../support/fixtures";
import { signIn } from "../../support/sign-in";

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
    test.fail(
        true,
        "Known bug: a disabled user still signs in - login answers 200 with a session, and only the calls " +
            "after it say 'User is unavailable'. Remove this once sign-in refuses them.",
    );
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
