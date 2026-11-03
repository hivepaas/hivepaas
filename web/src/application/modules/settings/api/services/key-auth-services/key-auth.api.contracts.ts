import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingKeyAuth } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type KeyAuth_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type KeyAuth_FindManyPaginated_Res = ApiResponsePaginated<SettingKeyAuth>;

export type KeyAuth_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type KeyAuth_FindOneById_Res = ApiResponseBase<SettingKeyAuth>;

export type KeyAuth_CreateOne_Payload = {
    inheritable: boolean;
    default: boolean;
    name: string;
    keyId: string;
    secretKey: string;
};
export type KeyAuth_CreateOne_Req = ApiRequestBase<{
    payload: KeyAuth_CreateOne_Payload;
}>;
export type KeyAuth_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type KeyAuth_UpdateOne_Payload = {
    updateVer: number;
    inheritable: boolean;
    default: boolean;
    name: string;
    keyId: string;
    secretKey: string;
};
export type KeyAuth_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: KeyAuth_UpdateOne_Payload;
}>;
export type KeyAuth_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type KeyAuth_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    inheritable?: boolean;
    default?: boolean;
};
export type KeyAuth_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: KeyAuth_UpdateStatus_Payload;
}>;
export type KeyAuth_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type KeyAuth_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type KeyAuth_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
