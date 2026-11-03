import { type ApiResponsePaginated } from "@infrastructure/api";

export const DEFAULT_PAGINATED_DATA: ApiResponsePaginated<never> = {
    data: [],
    meta: { page: { limit: 10, offset: 0, total: 0 } },
};
