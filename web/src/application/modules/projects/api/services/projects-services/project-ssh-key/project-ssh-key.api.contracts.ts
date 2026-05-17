import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    SSHKey_CreateOne_Payload,
    SSHKey_UpdateMeta_Payload,
    SSHKey_UpdateOne_Payload,
} from "~/settings/api/services";
import type { SettingSSHKey } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectSSHKey_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectSSHKey_FindManyPaginated_Res = ApiResponsePaginated<SettingSSHKey>;

export type ProjectSSHKey_FindOneById_Req = ApiRequestBase<{ projectID: string; id: string }>;

export type ProjectSSHKey_FindOneById_Res = ApiResponseBase<SettingSSHKey>;

export type ProjectSSHKey_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: SSHKey_CreateOne_Payload;
}>;

export type ProjectSSHKey_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectSSHKey_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: SSHKey_UpdateOne_Payload;
}>;

export type ProjectSSHKey_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectSSHKey_UpdateMeta_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: SSHKey_UpdateMeta_Payload;
}>;

export type ProjectSSHKey_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectSSHKey_DeleteOne_Req = ApiRequestBase<{ projectID: string; id: string }>;

export type ProjectSSHKey_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
