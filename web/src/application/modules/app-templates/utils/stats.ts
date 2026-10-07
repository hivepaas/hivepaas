import type { AppTemplateSort, AppTemplateSummary } from "../api";

/** How long a template is marked new after it joins the catalog. */
const NEW_FOR_DAYS = 14;
const DAY_MS = 24 * 60 * 60 * 1000;

/** The orders the store offers, A to Z first: it is the default. */
export const TEMPLATE_SORTS: { value: AppTemplateSort; label: string }[] = [
    { value: "name", label: "A–Z" },
    { value: "-stars", label: "Popular" },
    { value: "-trending", label: "Trending" },
    { value: "-added", label: "New" },
];

/** Whether the template joined the catalog in the last NEW_FOR_DAYS days. */
export function isNewTemplate(template: Pick<AppTemplateSummary, "added">, now: number = Date.now()): boolean {
    if (!template.added) return false;
    const added = Date.parse(`${template.added}T00:00:00Z`);
    return !Number.isNaN(added) && now - added >= 0 && now - added < NEW_FOR_DAYS * DAY_MS;
}

const compact = new Intl.NumberFormat("en", { notation: "compact", maximumFractionDigits: 1 });

/** 7812 as 7.8K: a card has room for the size of a number, not its digits. */
export function formatStars(stars: number): string {
    return compact.format(stars);
}
