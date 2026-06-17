import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    AcmeDnsProvider_CreateOne_Payload,
    AcmeDnsProvider_UpdateOne_Payload,
    AcmeDnsProvider_UpdateStatus_Payload,
} from "~/settings/api/services/acme-dns-provider-services";
import type { SettingAcmeDnsProvider } from "~/settings/domain";

import type { EAcmeDnsProviderKind } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectAcmeDnsProvider_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    kind?: EAcmeDnsProviderKind;
}>;
export type ProjectAcmeDnsProvider_FindManyPaginated_Res = ApiResponsePaginated<SettingAcmeDnsProvider>;

export type ProjectAcmeDnsProvider_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectAcmeDnsProvider_FindOneById_Res = ApiResponseBase<SettingAcmeDnsProvider>;

export type ProjectAcmeDnsProvider_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: AcmeDnsProvider_CreateOne_Payload;
}>;
export type ProjectAcmeDnsProvider_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type ProjectAcmeDnsProvider_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: AcmeDnsProvider_UpdateOne_Payload;
}>;
export type ProjectAcmeDnsProvider_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectAcmeDnsProvider_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    id: string;
    payload: AcmeDnsProvider_UpdateStatus_Payload;
}>;
export type ProjectAcmeDnsProvider_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type ProjectAcmeDnsProvider_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;
export type ProjectAcmeDnsProvider_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
