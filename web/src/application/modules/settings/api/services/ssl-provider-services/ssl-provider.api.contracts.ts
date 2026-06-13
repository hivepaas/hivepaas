import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingSslProvider } from "~/settings/domain";

import type { ESettingStatus, ESslKeyType, ESslProviderKind } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type SslProviderLetsEncrypt_Payload = Record<string, never>;

export type SslProviderEab_Payload = {
    eabKid: string;
    eabHmacKey: string;
};

export type SslProvider_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    name: string;
    kind: ESslProviderKind;
    email: string;
    defaultKeyType?: ESslKeyType;
    letsEncrypt: SslProviderLetsEncrypt_Payload | null;
    zeroSSL: SslProviderEab_Payload | null;
    googleTS: SslProviderEab_Payload | null;
};

export type SslProvider_UpdateOne_Payload = SslProvider_CreateOne_Payload & {
    updateVer: number;
};

export type SslProvider_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type SslProvider_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;
export type SslProvider_FindManyPaginated_Res = ApiResponsePaginated<SettingSslProvider>;

export type SslProvider_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type SslProvider_FindOneById_Res = ApiResponseBase<SettingSslProvider>;

export type SslProvider_CreateOne_Req = ApiRequestBase<{
    payload: SslProvider_CreateOne_Payload;
}>;
export type SslProvider_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type SslProvider_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: SslProvider_UpdateOne_Payload;
}>;
export type SslProvider_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SslProvider_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: SslProvider_UpdateStatus_Payload;
}>;
export type SslProvider_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type SslProvider_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type SslProvider_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
