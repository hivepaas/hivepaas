import type { Page } from "@playwright/test";

import { deleteSettingsNamed } from "../../support/api";
import { e2eName, expect, test } from "../../support/fixtures";

// The integrations whose saving calls nothing outside HivePaaS, each with what
// its form needs beyond a name, and the secret its edit form must not show.
const KINDS: {
    path: string;
    apiKind: string;
    noun: string;
    fill: (field: (name: RegExp) => ReturnType<Page["getByRole"]>, page: Page) => Promise<void>;
    secret?: RegExp;
}[] = [
    {
        path: "basic-auth",
        apiKind: "basic-auth",
        noun: "basic auth",
        fill: async field => {
            await field(/^Username/)
                .getByRole("textbox")
                .fill("e2e-user");
            await field(/^Password/)
                .getByRole("textbox")
                .fill("e2e-password");
        },
        secret: /^Password/,
    },
    {
        path: "key-auth",
        apiKind: "key-auth",
        noun: "key auth",
        fill: async field => {
            await field(/^Key ID/)
                .getByRole("textbox")
                .fill("e2e-key-id");
            await field(/^Secret Key/)
                .getByRole("textbox")
                .fill("e2e-secret");
        },
        secret: /^Secret Key/,
    },
    {
        path: "ssh-keys",
        apiKind: "ssh-keys",
        noun: "SSH key",
        fill: async (field, page) => {
            await page.getByRole("button", { name: "Generate" }).click();
            await expect(field(/^Public Key/).getByRole("textbox")).toHaveValue(/^ssh-ed25519 /);
        },
    },
    {
        path: "webhooks",
        apiKind: "repo-webhooks",
        noun: "webhook",
        // Its secret is generated when left empty.
        fill: async () => {},
    },
];

const capitalized = (noun: string) => noun.charAt(0).toUpperCase() + noun.slice(1);

for (const kind of KINDS) {
    test(`a ${kind.noun} is made, renamed and deleted from its screens`, async ({ page, api, cleanup }) => {
        const name = e2eName(kind.path);
        const renamed = `${name}-renamed`;
        cleanup(async () => {
            await deleteSettingsNamed(api, kind.apiKind, name);
            await deleteSettingsNamed(api, kind.apiKind, renamed);
        });
        const field = (label: RegExp) => page.getByRole("group", { name: label });
        const Noun = capitalized(kind.noun);

        await page.goto(`/integrations/${kind.path}/create/`);
        await field(/^Name/).getByRole("textbox").fill(name);
        await kind.fill(field, page);
        await page.getByRole("button", { name: "Save" }).click();

        await expect(page.getByText(`${Noun} created successfully`)).toBeVisible();
        await page.goto(`/integrations/${kind.path}/`);
        const row = page.getByRole("row").filter({ hasText: name });
        await expect(row).toBeVisible();

        await row.getByRole("button", { name: `Edit ${kind.noun}` }).click();
        if (kind.secret) {
            await expect(field(kind.secret).getByRole("textbox")).toHaveValue("********");
        }
        await field(/^Name/).getByRole("textbox").fill(renamed);
        await page.getByRole("button", { name: "Save" }).click();

        await expect(page.getByText(`${Noun} updated successfully`)).toBeVisible();
        const renamedRow = page.getByRole("row").filter({ hasText: renamed });
        await expect(renamedRow).toBeVisible();

        await renamedRow.getByRole("button", { name: "Actions menu" }).click();
        await page.getByRole("menu").getByRole("button", { name: "Delete" }).click();
        // The confirmation has no name of its own: found by its heading.
        const dialog = page
            .getByRole("dialog")
            .filter({ has: page.getByRole("heading", { name: `Delete ${kind.noun}` }) });
        await dialog.getByRole("button", { name: "Delete" }).click();

        await expect(page.getByText(`${Noun} deleted successfully`)).toBeVisible();
        await expect(page.getByRole("row").filter({ hasText: renamed })).toHaveCount(0);
    });
}
