import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    GithubApp_BeginManifestFlow_Payload,
    GithubApp_BeginManifestFlow_Res,
    GithubApp_BeginReprovision_Payload,
    GithubApp_BeginReprovision_Res,
    GithubApp_CreateOne_Payload,
    GithubApp_UpdateOne_Payload,
    GithubApp_UpdateStatus_Payload,
} from "~/settings/api/services";
import type { SettingGithubApp } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectGithubApp_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type ProjectGithubApp_FindManyPaginated_Res = ApiResponsePaginated<SettingGithubApp>;

export type ProjectGithubApp_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectGithubApp_FindOneById_Res = ApiResponseBase<SettingGithubApp>;

export type ProjectGithubApp_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: GithubApp_CreateOne_Payload;
}>;
export type ProjectGithubApp_CreateOne_Res = ApiResponseBase<{
    id: string;
    callbackURL?: string;
}>;

export type ProjectGithubApp_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: GithubApp_UpdateOne_Payload;
}>;
export type ProjectGithubApp_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectGithubApp_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: GithubApp_UpdateStatus_Payload;
}>;
export type ProjectGithubApp_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectGithubApp_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectGithubApp_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectGithubApp_BeginManifestFlow_Req = ApiRequestBase<{
    projectID: string;
    payload: GithubApp_BeginManifestFlow_Payload;
}>;
export type ProjectGithubApp_BeginManifestFlow_Res = GithubApp_BeginManifestFlow_Res;

export type ProjectGithubApp_BeginReprovision_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: GithubApp_BeginReprovision_Payload;
}>;
export type ProjectGithubApp_BeginReprovision_Res = GithubApp_BeginReprovision_Res;
