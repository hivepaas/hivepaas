export const SET_MANAGER_NODES_PAGE_SIZE = 50;

export type SelectorPagination = {
    page: number;
    size: number;
};

export function createSelectorPagination(page: number): SelectorPagination {
    return {
        page,
        size: SET_MANAGER_NODES_PAGE_SIZE,
    };
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
