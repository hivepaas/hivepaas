import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type ProjectConfigFile } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many project config files paginated
 */
export type ProjectConfigFiles_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectConfigFiles_FindManyPaginated_Res = ApiResponsePaginated<ProjectConfigFile>;

/**
 * Create project config file
 */
export type ProjectConfigFiles_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    name: string;
    content: string;
    base64: boolean;
    inheritable: boolean;
}>;

export type ProjectConfigFiles_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Find one project config file by id
 */
export type ProjectConfigFiles_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    configFileID: string;
}>;

export type ProjectConfigFiles_FindOneById_Res = ApiResponseBase<ProjectConfigFile>;

/**
 * Delete project config file
 */
export type ProjectConfigFiles_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    configFileID: string;
}>;

export type ProjectConfigFiles_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update project config file. An empty content keeps the current one.
 */
export type ProjectConfigFiles_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    configFileID: string;
    updateVer: number;
    name: string;
    content?: string;
    base64: boolean;
    inheritable: boolean;
}>;

export type ProjectConfigFiles_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;
