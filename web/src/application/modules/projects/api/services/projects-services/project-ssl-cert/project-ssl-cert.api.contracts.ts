import type { SettingSslCert } from "~/settings/domain";
import type {
    SslCert_CreateOne_Payload,
    SslCert_UpdateOne_Payload,
    SslCert_UpdateStatus_Payload,
} from "~/settings/api/services/ssl-cert-services";

import type { PaginationState, SortingState } from "@infrastructure/data";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectSslCert_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    domain?: string;
}>;

export type ProjectSslCert_FindManyPaginated_Res = ApiResponsePaginated<SettingSslCert>;

export type ProjectSslCert_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectSslCert_FindOneById_Res = ApiResponseBase<SettingSslCert>;

export type ProjectSslCert_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: SslCert_CreateOne_Payload;
}>;

export type ProjectSslCert_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectSslCert_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: SslCert_UpdateOne_Payload;
}>;

export type ProjectSslCert_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectSslCert_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: SslCert_UpdateStatus_Payload;
}>;

export type ProjectSslCert_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectSslCert_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectSslCert_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
