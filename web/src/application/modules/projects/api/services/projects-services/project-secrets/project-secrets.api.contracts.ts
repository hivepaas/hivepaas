import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectSecret } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many project secrets paginated
 */
export type ProjectSecrets_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectSecrets_FindManyPaginated_Res = ApiResponsePaginated<ProjectSecret>;

/**
 * Create project secret
 */
export type ProjectSecrets_CreateOne_Req = ApiRequestBase<
    {
        projectID: string;
    } & Pick<ProjectSecret, "name" | "key">
>;

export type ProjectSecrets_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Find one project secret by id
 */
export type ProjectSecrets_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    secretID: string;
}>;

export type ProjectSecrets_FindOneById_Res = ApiResponseBase<ProjectSecret>;

/**
 * Delete project secret
 */
export type ProjectSecrets_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    secretID: string;
}>;

export type ProjectSecrets_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project secret
 */
export type ProjectSecrets_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        secretID: string;
        updateVer: number;
    } & Partial<Omit<ProjectSecret, "id" | "key" | "createdAt" | "updatedAt" | "updateVer">>
>;

export type ProjectSecrets_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;
