import type { Page } from "@playwright/test";

// signIn fills the sign-in form and sends it, as a person does.
export async function signIn(page: Page, username: string, password: string): Promise<void> {
    await page.getByLabel("Username", { exact: true }).fill(username);
    await page.getByLabel("Password", { exact: true }).fill(password);
    await page.getByRole("button", { name: "Sign in" }).click();
}
