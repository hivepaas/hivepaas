import { type FilterByState, type PaginationState, type SortingState } from "@infrastructure/data";

interface Search {
    search: string;
}

interface Context {
    resourceContext: string;
}

interface Sorting {
    sort: string[] | string;
}

interface Pagination {
    pageLimit: number;
    pageOffset: number;
}

type QueryValue = string | number | boolean;
type QueryValueArray = (QueryValue | null | undefined)[];
type QueryObject = Record<string, QueryValue>;
type QueryObjectInput = Record<string, QueryValue | QueryValueArray>;

function serializeQueryValue(value: QueryValue | QueryValueArray): QueryValue | undefined {
    if (!Array.isArray(value)) {
        return value;
    }

    const values = value.filter((item): item is QueryValue => item !== undefined && item !== null && item !== "");

    if (values.length === 0) {
        return undefined;
    }

    if (values.length === 1) {
        return values[0];
    }

    return values.join(",");
}

function serializeQueryObject(query?: QueryObjectInput | null): QueryObject {
    if (!query) {
        return {};
    }

    return Object.entries(query).reduce<QueryObject>((acc, [key, value]) => {
        const serializedValue = serializeQueryValue(value);

        if (serializedValue === undefined) {
            return acc;
        }

        return {
            ...acc,
            [key]: serializedValue,
        };
    }, {});
}

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
            search,
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

        if (sorting.length === 1) {
            const [sort] = sorting;
            this.#sorting = {
                ...this.#sorting,
                sort: `${sort?.desc ? "-" : ""}${sort?.id}`,
            };
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
            pageLimit: size,
            pageOffset: (page - 1) * size,
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
            ...serializeQueryObject(sorting ? { sort: sorting.sort } : null),
            ...search,
            ...context,
            ...serializeQueryObject(filterBy),
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
