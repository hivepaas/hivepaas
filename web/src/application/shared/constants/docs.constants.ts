/** Where HivePaaS's documentation is. */
export const DOCS_URL = "https://docs.hivepaas.com";

/** The docs' pages on plans and license keys, which System > License and the plan-limit dialog lead to. */
export const LICENSE_DOCS = {
    /** What each plan allows, and what counts towards each limit. */
    plansAndLimits: `${DOCS_URL}/plans-and-licenses/plans-and-limits`,
    /** Finding the installation ID, buying a key, and pasting it. */
    installAKey: `${DOCS_URL}/plans-and-licenses/install-a-license-key`,
    /** The daily check for a newer key: exactly what it sends, and how to turn it off. */
    renewalAndPrivacy: `${DOCS_URL}/plans-and-licenses/renewal-and-privacy`,
} as const;
