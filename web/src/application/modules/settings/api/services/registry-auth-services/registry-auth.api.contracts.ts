import type { ESettingStatus } from "@application/shared/enums";
import type { SettingRegistryAuth } from "~/settings/domain";

import type { PaginationState, SortingState } from "@infrastructure/data";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type RegistryAuth_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type RegistryAuth_FindManyPaginated_Res = ApiResponsePaginated<SettingRegistryAuth>;

export type RegistryAuth_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;

export type RegistryAuth_FindOneById_Res = ApiResponseBase<SettingRegistryAuth>;

export type RegistryAuth_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    address: string;
    username: string;
    password: string;
    readonly: boolean;
};

export type RegistryAuth_CreateOne_Req = ApiRequestBase<{
    payload: RegistryAuth_CreateOne_Payload;
}>;

export type RegistryAuth_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type RegistryAuth_UpdateOne_Payload = {
    updateVer: number;
    availableInProjects: boolean;
    default: boolean;
    name: string;
    address: string;
    username: string;
    password: string;
    readonly: boolean;
};

export type RegistryAuth_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: RegistryAuth_UpdateOne_Payload;
}>;

export type RegistryAuth_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type RegistryAuth_UpdateMeta_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type RegistryAuth_UpdateMeta_Req = ApiRequestBase<{
    id: string;
    payload: RegistryAuth_UpdateMeta_Payload;
}>;

export type RegistryAuth_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type RegistryAuth_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;

export type RegistryAuth_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type RegistryAuth_TestConn_Payload = {
    name: string;
    address: string;
    username: string;
    password: string;
    readonly: boolean;
};

export type RegistryAuth_TestConn_Req = ApiRequestBase<{
    payload: RegistryAuth_TestConn_Payload;
}>;

export type RegistryAuth_TestConn_Res = ApiResponseBase<{ type: "success" }>;
