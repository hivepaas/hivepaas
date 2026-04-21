import type { ESettingStatus } from "@application/shared/enums";
import type { SettingBasicAuth } from "~/settings/domain";

import type { PaginationState, SortingState } from "@infrastructure/data";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type BasicAuth_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type BasicAuth_FindManyPaginated_Res = ApiResponsePaginated<SettingBasicAuth>;

export type BasicAuth_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type BasicAuth_FindOneById_Res = ApiResponseBase<SettingBasicAuth>;

export type BasicAuth_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    username: string;
    password: string;
};
export type BasicAuth_CreateOne_Req = ApiRequestBase<{
    payload: BasicAuth_CreateOne_Payload;
}>;
export type BasicAuth_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type BasicAuth_UpdateOne_Payload = {
    updateVer: number;
    availableInProjects: boolean;
    default: boolean;
    name: string;
    username: string;
    password: string;
};
export type BasicAuth_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: BasicAuth_UpdateOne_Payload;
}>;
export type BasicAuth_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type BasicAuth_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};
export type BasicAuth_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: BasicAuth_UpdateStatus_Payload;
}>;
export type BasicAuth_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type BasicAuth_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type BasicAuth_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
