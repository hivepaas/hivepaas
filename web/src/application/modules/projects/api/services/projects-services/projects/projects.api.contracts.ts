import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectBaseEntity } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many projects paginated
 */
export type Projects_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Projects_FindManyPaginated_Res = ApiResponsePaginated<ProjectBaseEntity>;
