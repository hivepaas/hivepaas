import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type AppConfigFile } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many app config files paginated
 */
export type AppConfigFiles_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppConfigFiles_FindManyPaginated_Res = ApiResponsePaginated<AppConfigFile>;

/**
 * Find one app config file by id
 */
export type AppConfigFiles_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    configFileID: string;
}>;

export type AppConfigFiles_FindOneById_Res = ApiResponseBase<AppConfigFile>;

/**
 * Get app config file download token
 */
export type AppConfigFiles_GetDownloadToken_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    configFileID: string;
}>;

export type AppConfigFiles_GetDownloadToken_Res = ApiResponseBase<{
    token: string;
}>;

/**
 * Build app config file download URL
 */
export type AppConfigFiles_BuildDownloadUrl_Req = {
    projectID: string;
    appID: string;
    configFileID: string;
    token: string;
    viewInline: true;
};

/**
 * Create app config file
 */
export type AppConfigFiles_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    name: string;
    content: string;
    base64: boolean;
    swarmRef?: AppConfigFile["swarmRef"];
}>;

export type AppConfigFiles_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Delete app config file
 */
export type AppConfigFiles_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    configFileID: string;
}>;

export type AppConfigFiles_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update app config file
 */
export type AppConfigFiles_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        appID: string;
        configFileID: string;
        updateVer: number;
    } & Partial<Omit<AppConfigFile, "id" | "createdAt" | "updatedAt" | "updateVer" | "inherited">>
>;

export type AppConfigFiles_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;
