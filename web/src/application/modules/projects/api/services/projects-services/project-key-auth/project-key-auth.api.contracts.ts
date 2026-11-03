import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    KeyAuth_CreateOne_Payload,
    KeyAuth_UpdateOne_Payload,
    KeyAuth_UpdateStatus_Payload,
} from "~/settings/api/services/key-auth-services";
import type { SettingKeyAuth } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectKeyAuth_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type ProjectKeyAuth_FindManyPaginated_Res = ApiResponsePaginated<SettingKeyAuth>;

export type ProjectKeyAuth_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
}>;
export type ProjectKeyAuth_FindOneById_Res = ApiResponseBase<SettingKeyAuth>;

export type ProjectKeyAuth_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    payload: KeyAuth_CreateOne_Payload;
}>;
export type ProjectKeyAuth_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectKeyAuth_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
    payload: KeyAuth_UpdateOne_Payload;
}>;
export type ProjectKeyAuth_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectKeyAuth_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
    payload: KeyAuth_UpdateStatus_Payload;
}>;
export type ProjectKeyAuth_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectKeyAuth_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    env?: string;
    id: string;
}>;
export type ProjectKeyAuth_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
