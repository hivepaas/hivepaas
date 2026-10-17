import { env } from "../../support/env";
import { expect, test } from "../../support/fixtures";
import { signIn } from "../../support/sign-in";

test.describe("signed out", () => {
    test.use({ storageState: { cookies: [], origins: [] } });

    test("a user signs in with a username and password", async ({ page }) => {
        await page.goto("/");
        await expect(page).toHaveURL(/\/auth\/sign-in\/(\?.*)?$/);

        await signIn(page, env.username, env.password);

        await expect(page).toHaveURL(/\/home\/$/);
        await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    });

    // A sign-in with a provider comes back to this page with the error's code,
    // never a JSON error the browser would stop on.
    test("a failed sign-in with a provider says why, on the sign-in page", async ({ page }) => {
        await page.goto("/auth/sign-in/?ssoError=ERR_OAUTH_EMAIL_NOT_VERIFIED");

        await expect(page.getByRole("alert")).toContainText("does not confirm that your email address is verified");
    });

    test("a sign-in with no such provider is sent back to the sign-in page", async ({ playwright }) => {
        const request = await playwright.request.newContext({ baseURL: `${env.baseURL}/api/` });
        const res = await request.get("auth/sso/e2e-no-such-provider", { maxRedirects: 0 });

        expect(res.status()).toBe(302);
        expect(res.headers().location).toMatch(/\/auth\/sign-in\?ssoError=ERR_NOT_FOUND$/);
        await request.dispose();
    });

    // An unknown user, never a real one with a wrong password: failed attempts
    // block an account for 15 minutes.
    test("an unknown user is turned away, on the sign-in page", async ({ page }) => {
        await page.goto("/");

        await signIn(page, `e2e-nobody-${Date.now()}`, "not-the-password");

        await expect(page.getByText("Email or password is incorrect")).toBeVisible();
        await expect(page).toHaveURL(/\/auth\/sign-in\/(\?.*)?$/);
    });
});

test("a saved session opens the dashboard without signing in", async ({ page }) => {
    await page.goto("/");

    await expect(page).toHaveURL(/\/home\/$/);
    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
});

test("the API fixture is signed in as the tests' user", async ({ api }) => {
    const res = await api.get("sessions/me");

    expect(res.ok()).toBe(true);
    const { data } = (await res.json()) as { data: { user: { username: string; email: string } } };
    expect([data.user.username, data.user.email]).toContain(env.username);
});
