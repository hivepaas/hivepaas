import { GIT_SELECTOR_PAGE_SIZE } from "./git-repository.utils";

export type SelectorPagination = {
    page: number;
    size: number;
};

export function createSelectorPagination(page: number): SelectorPagination {
    return {
        page,
        size: GIT_SELECTOR_PAGE_SIZE,
    };
}

export function includesSelectorSearch(value: string | number | null | undefined, search: string): boolean {
    return String(value ?? "")
        .toLowerCase()
        .includes(search);
}

export function appendUniqueByKey<T>(current: T[], next: T[], getKey: (item: T) => string): T[] {
    const keys = new Set(current.map(getKey));
    const merged = [...current];

    for (const item of next) {
        const key = getKey(item);

        if (keys.has(key)) {
            continue;
        }

        keys.add(key);
        merged.push(item);
    }

    return merged;
}
