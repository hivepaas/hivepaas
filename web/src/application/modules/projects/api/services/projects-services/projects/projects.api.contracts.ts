import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectBaseEntity } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many projects paginated
 */
export type Projects_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type Projects_FindManyPaginated_Res = ApiResponsePaginated<ProjectBaseEntity>;

/**
 * Create project
 */
export type Projects_CreateOne_Req = ApiRequestBase<Pick<ProjectBaseEntity, "name" | "note" | "tags">>;

export type Projects_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Delete project
 */
export type Projects_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
}>;

export type Projects_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;