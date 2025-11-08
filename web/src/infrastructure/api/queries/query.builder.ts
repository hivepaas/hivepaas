import { type FilterByState, type PaginationState, type SortingState } from "@infrastructure/data";

interface Search {
    "filter[search]": string;
}

interface Context {
    resourceContext: string;
}

interface Sorting {
    sort: string[];
}

interface Pagination {
    "page[limit]": number;
    "page[offset]": number;
}

type QueryObject = Record<string, string | number | (string | number)[]>;

class Builder {
    #pagination: Pagination | null = null;
    #sorting: Sorting | null = null;
    #search: Search | null = null;
    #context: Context | null = null;
    #filterBy: FilterByState | null = null;

    /**
     * Add filter to the query
     */
    filterBy(filterBy?: Partial<FilterByState>): this {
        if (!filterBy) {
            return this;
        }

        this.#filterBy = {
            ...this.#filterBy,

            ...Object.entries(filterBy).reduce((acc, [key, val]) => {
                if (!val) {
                    return acc;
                }

                const values = val.filter(v => v !== undefined && v !== null && v !== "");

                if (values.length === 0) {
                    return acc;
                }

                return {
                    ...acc,
                    [key]: values,
                };
            }, {}),
        };

        return this;
    }

    /**
     * Add resource context to a query
     */
    context(context?: string): this {
        if (!context) {
            return this;
        }

        this.#context = {
            resourceContext: context,
        };

        return this;
    }

    /**
     * Add search to a query
     */
    search(search?: string): this {
        if (!search) {
            return this;
        }

        this.#search = {
            "filter[search]": search,
        };

        return this;
    }

    /**
     * Add sort to a query
     */
    sorting(sorting?: SortingState): this {
        if (!sorting) {
            return this;
        }

        this.#sorting = {
            ...this.#sorting,

            sort: sorting.map(({ id, desc }) => `${desc ? "-" : ""}${id}`),
        };

        return this;
    }

    /**
     * Add pagination to a query
     */
    pagination(pagination?: PaginationState): this {
        if (!pagination) {
            return this;
        }

        const { page, size } = pagination;

        if (page < 1) {
            throw new Error("Page must be greater than 0");
        }

        if (size < 1) {
            throw new Error("Limit must be greater than 0");
        }

        this.#pagination = {
            "page[limit]": size,
            "page[offset]": (page - 1) * size,
        };

        return this;
    }

    /**
     * Build query object
     */
    build(): QueryObject {
        const pagination = this.#pagination;
        const sorting = this.#sorting;
        const search = this.#search;
        const context = this.#context;
        const filterBy = this.#filterBy;

        this.clear();

        return {
            ...pagination,
            ...sorting,
            ...search,
            ...context,
            ...filterBy,
        };
    }

    /**
     * Clear query
     */
    private clear(): void {
        this.#pagination = null;
        this.#sorting = null;
        this.#search = null;
        this.#context = null;
        this.#filterBy = null;
    }
}

export class QueryBuilder {
    getInstance(): Builder {
        return new Builder();
    }
}
