import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingBasicAuth } from "~/settings/domain";
import type {
    BasicAuth_CreateOne_Payload,
    BasicAuth_UpdateOne_Payload,
    BasicAuth_UpdateStatus_Payload,
} from "~/settings/api/services/basic-auth-services";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectBasicAuth_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type ProjectBasicAuth_FindManyPaginated_Res = ApiResponsePaginated<SettingBasicAuth>;

export type ProjectBasicAuth_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectBasicAuth_FindOneById_Res = ApiResponseBase<SettingBasicAuth>;

export type ProjectBasicAuth_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: BasicAuth_CreateOne_Payload;
}>;
export type ProjectBasicAuth_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectBasicAuth_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: BasicAuth_UpdateOne_Payload;
}>;
export type ProjectBasicAuth_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectBasicAuth_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: BasicAuth_UpdateStatus_Payload;
}>;
export type ProjectBasicAuth_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectBasicAuth_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectBasicAuth_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
