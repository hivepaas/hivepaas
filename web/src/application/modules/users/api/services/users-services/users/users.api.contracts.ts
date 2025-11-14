import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type UserBase } from "~/users/domain";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many users paginated
 */
export type Users_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Users_FindManyPaginated_Res = ApiResponsePaginated<UserBase>;
