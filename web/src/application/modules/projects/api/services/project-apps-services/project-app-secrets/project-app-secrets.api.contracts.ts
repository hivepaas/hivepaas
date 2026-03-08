import { type PaginationState, type SortingState } from "@infrastructure/data";
import { type AppSecret } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many app secrets paginated
 */
export type AppSecrets_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppSecrets_FindManyPaginated_Res = ApiResponsePaginated<AppSecret>;

/**
 * Find one app secret by id
 */
export type AppSecrets_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    secretID: string;
}>;

export type AppSecrets_FindOneById_Res = ApiResponseBase<AppSecret>;

/**
 * Create app secret
 */
export type AppSecrets_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    name: string;
    value: string;
}>;

export type AppSecrets_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

/**
 * Delete app secret
 */
export type AppSecrets_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    secretID: string;
}>;

export type AppSecrets_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

/**
 * Update app secret
 */
export type AppSecrets_UpdateOne_Req = ApiRequestBase<
    {
        projectID: string;
        appID: string;
        secretID: string;
        updateVer: number;
    } & Partial<Omit<AppSecret, "id" | "key" | "createdAt" | "updatedAt" | "updateVer" | "inherited">> & {
            value?: string;
        }
>;

export type AppSecrets_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;
