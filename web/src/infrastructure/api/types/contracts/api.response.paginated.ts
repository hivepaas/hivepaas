import { type Prettify } from "@infrastructure/utility-types";

interface Meta {
    code?: string;
    message?: string;
}

interface PagingMeta extends Meta {
    page: {
        limit: number;
        offset: number;
        total: number;
    };
}

export interface ApiResponsePaginated<TData, TMeta extends object = object> {
    data: TData[];
    meta: Prettify<PagingMeta & TMeta>;
}
