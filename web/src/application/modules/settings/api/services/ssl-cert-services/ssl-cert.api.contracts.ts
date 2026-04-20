import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingSslCert } from "~/settings/domain";

import type { ESettingStatus, ESslCertType, ESslKeyType } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

/** Mirrors `basedto.BaseEventNotificationReq` for create/update bodies. */
export type SslCert_Notification_Payload = {
    success: { id: string };
    successUseDefault: boolean;
    failure: { id: string };
    failureUseDefault: boolean;
};

export type SslCert_CreateOne_Payload = {
    availableInProjects: boolean;
    default: boolean;
    certType: ESslCertType;
    domain: string;
    certificate: string;
    privateKey: string;
    keyType: ESslKeyType;
    validPeriod: string;
    email: string;
    autoRenew: boolean;
    expireAt: Date;
    notifyFrom: Date;
    notification?: SslCert_Notification_Payload | null;
};

export type SslCert_UpdateOne_Payload = {
    updateVer: number;
    availableInProjects: boolean;
    default: boolean;
    certType: ESslCertType;
    domain: string;
    certificate: string;
    privateKey: string;
    keyType: ESslKeyType;
    validPeriod: string;
    email: string;
    autoRenew: boolean;
    expireAt: Date;
    notifyFrom: Date;
    notification?: SslCert_Notification_Payload | null;
};

export type SslCert_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type SslCert_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    domain?: string;
}>;

export type SslCert_FindManyPaginated_Res = ApiResponsePaginated<SettingSslCert>;

export type SslCert_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;

export type SslCert_FindOneById_Res = ApiResponseBase<SettingSslCert>;

export type SslCert_CreateOne_Req = ApiRequestBase<{
    payload: SslCert_CreateOne_Payload;
}>;

export type SslCert_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type SslCert_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: SslCert_UpdateOne_Payload;
}>;

export type SslCert_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SslCert_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: SslCert_UpdateStatus_Payload;
}>;

export type SslCert_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type SslCert_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;

export type SslCert_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
