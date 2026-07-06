import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingSSHKey } from "~/settings/domain";

import type { ESSHKeyKind, ESSHKeyType, ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type SSHKey_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type SSHKey_FindManyPaginated_Res = ApiResponsePaginated<SettingSSHKey>;

export type SSHKey_FindOneById_Req = ApiRequestBase<{ id: string }>;

export type SSHKey_FindOneById_Res = ApiResponseBase<SettingSSHKey>;

export type SSHKey_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    kind: ESSHKeyKind;
    keyType: ESSHKeyType | "";
    publicKey: string;
    privateKey: string;
    passphrase: string;
};

export type SSHKey_CreateOne_Req = ApiRequestBase<{ payload: SSHKey_CreateOne_Payload }>;

export type SSHKey_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type SSHKey_UpdateOne_Payload = SSHKey_CreateOne_Payload & {
    updateVer: number;
};

export type SSHKey_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: SSHKey_UpdateOne_Payload;
}>;

export type SSHKey_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SSHKey_UpdateMeta_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type SSHKey_UpdateMeta_Req = ApiRequestBase<{
    id: string;
    payload: SSHKey_UpdateMeta_Payload;
}>;

export type SSHKey_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type SSHKey_DeleteOne_Req = ApiRequestBase<{ id: string }>;

export type SSHKey_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type SSHKey_Generate_Payload = {
    keyType: ESSHKeyType;
    passphrase?: string;
};

export type SSHKey_Generate_Req = ApiRequestBase<{ payload: SSHKey_Generate_Payload }>;

export type SSHKey_Generate_Res = ApiResponseBase<{
    keyType: ESSHKeyType;
    publicKey: string;
    privateKey: string;
}>;
