import type { APIRequestContext, Page } from "@playwright/test";

import { systemProject } from "../../support/api";
import { expect, test } from "../../support/fixtures";

// The store's orders, as the select names them, and as the API takes them.
const ORDERS: [label: string, sort: string][] = [
    ["A–Z", "name"],
    ["Popular", "-stars"],
    ["Trending", "-trending"],
    ["New", "-added"],
];

// How many cards are compared: the first page's first row or two, enough to
// tell one order from another.
const COMPARED = 8;

interface TemplateSummary {
    title: string;
}

// apiOrder is the titles of the first templates the API lists in that order,
// with the search, if any.
async function apiOrder(api: APIRequestContext, sort: string, search = ""): Promise<string[]> {
    const res = await api.get("app-templates", { params: { sort, search, pageLimit: COMPARED } });
    expect(res.ok(), `listing templates by ${sort}: ${res.status()}`).toBe(true);
    return ((await res.json()) as { data: TemplateSummary[] }).data.map(t => t.title);
}

// shownOrder is the titles of the first cards the store shows. A card is a
// button headed by its template's title.
async function shownOrder(page: Page, count: number): Promise<string[]> {
    return (await page.getByRole("button").getByRole("heading", { level: 3 }).allInnerTexts())
        .slice(0, count)
        .map(t => t.trim());
}

// The store lists its templates in the order chosen, as the API orders them,
// A to Z until another is chosen; a search keeps the order. Nothing is written:
// the store is browsed in the HivePaaS project, which every installation has.
test("the template store lists its templates in the order chosen, A to Z first", async ({ page, api }) => {
    const project = await systemProject(api);
    await page.goto(`/projects/${project.id}/app-templates/`);
    const order = page.getByRole("combobox", { name: "Order templates by" });
    await expect(order).toHaveText(/A–Z/);

    for (const [label, sort] of ORDERS) {
        if (label !== "A–Z") {
            await order.click();
            await page.getByRole("option", { name: label, exact: true }).click();
            await expect(order).toHaveText(new RegExp(label));
        }
        const expected = await apiOrder(api, sort);
        expect(expected.length, "the catalog has templates to order").toBeGreaterThan(1);
        await expect
            .poll(() => shownOrder(page, expected.length), { message: `the store ordered ${label}` })
            .toEqual(expected);
    }

    // A search narrows the list, in the order chosen - New, still.
    await page.getByRole("textbox", { name: /^Search templates/ }).fill("sql");
    const expected = await apiOrder(api, "-added", "sql");
    expect(expected.length, "a template matches sql").toBeGreaterThan(0);
    await expect
        .poll(() => shownOrder(page, expected.length), { message: "the search kept the order" })
        .toEqual(expected);
});
