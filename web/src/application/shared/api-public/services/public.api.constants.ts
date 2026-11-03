import type { PaginationState } from "@infrastructure/data";

/**
 * The pickers and filters list every item they may offer. Without a page asked
 * for, a list asks for the most the API returns in one page (PageLimitMax), as
 * the lists made for them used to by default.
 */
export const PUBLIC_LIST_ALL: PaginationState = { page: 1, size: 10_000 };

/** The status of a project, an app or a user that can be picked. */
export const PUBLIC_LIST_STATUS_ACTIVE = "active";
