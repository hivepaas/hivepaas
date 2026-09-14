/** What the text in a search box means. Plain and case-insensitive is the fast default. */
export interface SearchMode {
    isRegex: boolean;
    isCaseSensitive: boolean;
}

export const DEFAULT_SEARCH_MODE: SearchMode = { isRegex: false, isCaseSensitive: false };
