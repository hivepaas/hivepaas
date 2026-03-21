import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingRegistryAuth } from "~/settings/domain";
import type {
    RegistryAuth_CreateOne_Payload,
    RegistryAuth_UpdateMeta_Payload,
    RegistryAuth_UpdateOne_Payload,
} from "~/settings/api/services/registry-auth-services";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectRegistryAuth_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectRegistryAuth_FindManyPaginated_Res = ApiResponsePaginated<SettingRegistryAuth>;

export type ProjectRegistryAuth_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectRegistryAuth_FindOneById_Res = ApiResponseBase<SettingRegistryAuth>;

export type ProjectRegistryAuth_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: RegistryAuth_CreateOne_Payload;
}>;

export type ProjectRegistryAuth_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectRegistryAuth_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: RegistryAuth_UpdateOne_Payload;
}>;

export type ProjectRegistryAuth_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectRegistryAuth_UpdateMeta_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: RegistryAuth_UpdateMeta_Payload;
}>;

export type ProjectRegistryAuth_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectRegistryAuth_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectRegistryAuth_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
