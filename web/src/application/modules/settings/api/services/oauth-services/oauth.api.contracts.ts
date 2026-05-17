import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingOAuth } from "~/settings/domain";

import type { EOAuthKind, ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type OAuth_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type OAuth_FindManyPaginated_Res = ApiResponsePaginated<SettingOAuth>;

export type OAuth_FindOneById_Req = ApiRequestBase<{ id: string }>;

export type OAuth_FindOneById_Res = ApiResponseBase<SettingOAuth>;

export type OAuth_CreateOne_Payload = {
    default: boolean;
    kind: EOAuthKind;
    name: string;
    organization: string;
    clientId: string;
    clientSecret: string;
    authURL: string;
    tokenURL: string;
    profileURL: string;
    autoDiscoveryURL: string;
    scopes: string[];
};

export type OAuth_CreateOne_Req = ApiRequestBase<{ payload: OAuth_CreateOne_Payload }>;

export type OAuth_CreateOne_Res = ApiResponseBase<{ id: string; callbackURL?: string }>;

export type OAuth_UpdateOne_Payload = OAuth_CreateOne_Payload & {
    updateVer: number;
};

export type OAuth_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: OAuth_UpdateOne_Payload;
}>;

export type OAuth_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type OAuth_UpdateMeta_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    default?: boolean;
};

export type OAuth_UpdateMeta_Req = ApiRequestBase<{
    id: string;
    payload: OAuth_UpdateMeta_Payload;
}>;

export type OAuth_UpdateMeta_Res = ApiResponseBase<{ type: "success" }>;

export type OAuth_DeleteOne_Req = ApiRequestBase<{ id: string }>;

export type OAuth_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
