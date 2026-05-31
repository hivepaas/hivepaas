import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    RepoWebhook_CreateOne_Payload,
    RepoWebhook_CreateOne_Res,
    RepoWebhook_UpdateOne_Payload,
    RepoWebhook_UpdateStatus_Payload,
} from "~/settings/api/services";
import type { SettingRepoWebhook } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectRepoWebhook_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type ProjectRepoWebhook_FindManyPaginated_Res = ApiResponsePaginated<SettingRepoWebhook>;

export type ProjectRepoWebhook_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectRepoWebhook_FindOneById_Res = ApiResponseBase<SettingRepoWebhook>;

export type ProjectRepoWebhook_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: RepoWebhook_CreateOne_Payload;
}>;
export type ProjectRepoWebhook_CreateOne_Res = RepoWebhook_CreateOne_Res;

export type ProjectRepoWebhook_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: RepoWebhook_UpdateOne_Payload;
}>;
export type ProjectRepoWebhook_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectRepoWebhook_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: RepoWebhook_UpdateStatus_Payload;
}>;
export type ProjectRepoWebhook_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectRepoWebhook_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectRepoWebhook_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
