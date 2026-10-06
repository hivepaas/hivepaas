import { env } from "../../support/env";
import { expect, test } from "../../support/fixtures";
import { signIn } from "../../support/sign-in";

// The sign-in page, with the page asked for kept to come back to.
const SIGN_IN = /\/auth\/sign-in\/(\?.*)?$/;

test.describe("signed out", () => {
    // Signs in afresh, not with the shared session: logging out ends the
    // session it is done in.
    test.use({ storageState: { cookies: [], origins: [] } });

    test("logging out ends the session, and the dashboard asks to sign in again", async ({ page, api }) => {
        const me = await api.get("sessions/me");
        const { data } = (await me.json()) as { data: { user: { email: string } } };
        await page.goto("/");
        await signIn(page, env.username, env.password);
        await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();

        await page.getByRole("button", { name: data.user.email }).click();
        await page.getByRole("menuitem", { name: "Log out" }).click();

        await expect(page).toHaveURL(SIGN_IN);
        await page.goto("/home/");
        await expect(page).toHaveURL(SIGN_IN);
    });
});

test("an address that does not exist shows the not-found page", async ({ page }) => {
    await page.goto("/no-such-page/");

    await expect(page.getByRole("heading", { name: "Sorry, the page not found" })).toBeVisible();
    await expect(page.getByRole("button", { name: "Back to application" })).toBeVisible();
});
