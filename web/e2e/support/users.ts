import { type APIRequestContext, type Browser, type Page, request } from "@playwright/test";

import { type App, type Project, createApp, createProject, deleteProject, deleteUsersByEmail } from "./api";
import { env } from "./env";
import { type Cleanup, e2eName, expect, test } from "./fixtures";

// Members: invited by an admin, signed up from the link, signed in - on a page
// of their own, or to the API.

export const SIGNED_OUT = { baseURL: env.baseURL, storageState: { cookies: [], origins: [] } };
export const PASSWORD = "E2e-Passw0rd-member";

export interface Member {
    username: string;
    email: string;
}

// inviteAndSignUp has an admin invite a member to one project, by a link
// rather than an email, and the member sign up from it.
export async function inviteAndSignUp(page: Page, browser: Browser, label: string, project: Project): Promise<Member> {
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

// signedInAs is the API as the user signs in to it.
export async function signedInAs(username: string, password: string): Promise<APIRequestContext> {
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

interface UserDetails {
    id: string;
    username: string;
    email: string;
    fullName: string;
    role: string;
    status: string;
    securityOption: string;
    moduleAccesses?: { id: string; access: object }[] | null;
}

// grantMember has an admin give a member, as the member's page does, write to
// one environment of a project and the capabilities given - those left out
// taken away.
export async function grantMember(
    api: APIRequestContext,
    email: string,
    projectId: string,
    envId: string,
    capabilities: string[],
): Promise<void> {
    const list = await api.get("users", { params: { search: email } });
    const found = ((await list.json()) as { data: { id: string; email: string }[] }).data.find(u => u.email === email);
    expect(found, `the user ${email}`).toBeDefined();
    const current = await api.get(`users/${found?.id ?? ""}`, { params: { getAccesses: true } });
    const user = ((await current.json()) as { data: UserDetails }).data;
    const res = await api.put(`users/${user.id}`, {
        data: {
            username: user.username,
            email: user.email,
            fullName: user.fullName,
            role: user.role,
            status: user.status,
            securityOption: user.securityOption,
            moduleAccesses: (user.moduleAccesses ?? []).map(({ id, access }) => ({ id, access })),
            capabilities,
            projectAccesses: [
                {
                    project: { id: projectId },
                    envAccesses: [{ id: envId, access: { read: true, execute: true, write: true, delete: false } }],
                },
            ],
        },
    });
    expect(res.ok(), `granting ${email}: ${res.status()} ${await res.text()}`).toBe(true);
}

// What the member is refused answers 401, or 404 for what it may not know is
// there: anything else is the member reaching it.
export const REFUSED = [401, 403, 404];

export interface DevOnly {
    project: Project;
    dev: App;
    prod: App;
    member: Member;
}

// devOnly is a project with an app in development and one in production, and
// a member given development alone: to read, run and change, not to delete.
export async function devOnly(
    page: Page,
    browser: Browser,
    api: APIRequestContext,
    cleanup: Cleanup,
    label: string,
): Promise<DevOnly> {
    const project = await createProject(api, e2eName(label));
    cleanup(() => deleteProject(api, project.id));
    const dev = await createApp(api, project, "web");
    const prod = await createApp(api, project, "api", "production");
    cleanup(() => deleteUsersByEmail(api, `${e2eName(label)}@example.com`));
    const member = await inviteAndSignUp(page, browser, label, project);
    await grantMember(api, member.email, project.id, `${project.id}:dev`, []);
    return { project, dev, prod, member };
}
