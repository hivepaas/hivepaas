import { expect, test as setup } from "@playwright/test";

import { authFile, env } from "../support/env";
import { signIn } from "../support/sign-in";

// The session every other test starts from: signed in once, through the form,
// and kept with the token the dashboard stored and its refresh cookie.
setup("sign in", async ({ page }) => {
    await page.goto("/");
    await signIn(page, env.username, env.password);
    await expect(page.getByRole("heading", { name: "Welcome back" })).toBeVisible();
    await page.context().storageState({ path: authFile });
});
