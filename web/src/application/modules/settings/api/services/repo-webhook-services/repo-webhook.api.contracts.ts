import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingRepoWebhook } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type RepoWebhook_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type RepoWebhook_FindManyPaginated_Res = ApiResponsePaginated<SettingRepoWebhook>;

export type RepoWebhook_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type RepoWebhook_FindOneById_Res = ApiResponseBase<SettingRepoWebhook>;

export type RepoWebhook_BasePayload = {
    name: string;
    kind: string;
    secret: string;
};

export type RepoWebhook_CreateOne_Payload = RepoWebhook_BasePayload & {
    availableInProjects: boolean;
    default: boolean;
};
export type RepoWebhook_CreateOne_Req = ApiRequestBase<{
    payload: RepoWebhook_CreateOne_Payload;
}>;
export type RepoWebhook_CreateOne_Res = ApiResponseBase<{
    id: string;
    secret: string;
    webhookURL: string;
}>;

export type RepoWebhook_UpdateOne_Payload = RepoWebhook_CreateOne_Payload & {
    updateVer: number;
};
export type RepoWebhook_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: RepoWebhook_UpdateOne_Payload;
}>;
export type RepoWebhook_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type RepoWebhook_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};
export type RepoWebhook_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: RepoWebhook_UpdateStatus_Payload;
}>;
export type RepoWebhook_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type RepoWebhook_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type RepoWebhook_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
