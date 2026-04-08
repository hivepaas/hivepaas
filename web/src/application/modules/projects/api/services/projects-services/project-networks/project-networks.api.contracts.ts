import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectNetworkEntity } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many project networks paginated
 */
export type ProjectNetworks_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectNetworks_FindManyPaginated_Res = ApiResponsePaginated<ProjectNetworkEntity>;
