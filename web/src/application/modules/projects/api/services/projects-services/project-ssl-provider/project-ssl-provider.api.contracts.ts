import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    SslProvider_CreateOne_Payload,
    SslProvider_UpdateOne_Payload,
    SslProvider_UpdateStatus_Payload,
} from "~/settings/api/services/ssl-provider-services";
import type { SettingSslProvider } from "~/settings/domain";

import type { ESslProviderKind } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectSslProvider_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    kind?: ESslProviderKind;
}>;
export type ProjectSslProvider_FindManyPaginated_Res = ApiResponsePaginated<SettingSslProvider>;

export type ProjectSslProvider_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectSslProvider_FindOneById_Res = ApiResponseBase<SettingSslProvider>;

export type ProjectSslProvider_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: SslProvider_CreateOne_Payload;
}>;
export type ProjectSslProvider_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectSslProvider_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: SslProvider_UpdateOne_Payload;
}>;
export type ProjectSslProvider_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectSslProvider_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: SslProvider_UpdateStatus_Payload;
}>;
export type ProjectSslProvider_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectSslProvider_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectSslProvider_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
