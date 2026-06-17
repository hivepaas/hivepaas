import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    SettingAcmeDnsProvider,
    SettingAcmeDnsProviderAcmeDNS,
    SettingAcmeDnsProviderAzure,
    SettingAcmeDnsProviderBaiduCloud,
    SettingAcmeDnsProviderCloudflare,
    SettingAcmeDnsProviderDigitalOcean,
    SettingAcmeDnsProviderGCloud,
    SettingAcmeDnsProviderGoDaddy,
    SettingAcmeDnsProviderHetzner,
    SettingAcmeDnsProviderHuaweiCloud,
    SettingAcmeDnsProviderNamecheap,
    SettingAcmeDnsProviderRFC2136,
    SettingAcmeDnsProviderRoute53,
    SettingAcmeDnsProviderTencentCloud,
} from "~/settings/domain";

import type { EAcmeDnsProviderKind, ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AcmeDnsProvider_ConfigPayload = {
    kind: EAcmeDnsProviderKind;
    name: string;
    acmeDns: SettingAcmeDnsProviderAcmeDNS | null;
    azure: SettingAcmeDnsProviderAzure | null;
    baiduCloud: SettingAcmeDnsProviderBaiduCloud | null;
    cloudflare: SettingAcmeDnsProviderCloudflare | null;
    digitalOcean: SettingAcmeDnsProviderDigitalOcean | null;
    gCloud: SettingAcmeDnsProviderGCloud | null;
    goDaddy: SettingAcmeDnsProviderGoDaddy | null;
    hetzner: SettingAcmeDnsProviderHetzner | null;
    huaweiCloud: SettingAcmeDnsProviderHuaweiCloud | null;
    namecheap: SettingAcmeDnsProviderNamecheap | null;
    rfc2136: SettingAcmeDnsProviderRFC2136 | null;
    route53: SettingAcmeDnsProviderRoute53 | null;
    tencentCloud: SettingAcmeDnsProviderTencentCloud | null;
};

export type AcmeDnsProvider_CreateOne_Payload = AcmeDnsProvider_ConfigPayload & {
    availableInProjects: boolean;
    default: boolean;
};

export type AcmeDnsProvider_UpdateOne_Payload = AcmeDnsProvider_CreateOne_Payload & {
    updateVer: number;
};

export type AcmeDnsProvider_TestAccess_Payload = AcmeDnsProvider_ConfigPayload & {
    testDomain: string;
};

export type AcmeDnsProvider_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type AcmeDnsProvider_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    kind?: EAcmeDnsProviderKind;
}>;
export type AcmeDnsProvider_FindManyPaginated_Res = ApiResponsePaginated<SettingAcmeDnsProvider>;

export type AcmeDnsProvider_FindOneById_Req = ApiRequestBase<{
    id: string;
}>;
export type AcmeDnsProvider_FindOneById_Res = ApiResponseBase<SettingAcmeDnsProvider>;

export type AcmeDnsProvider_CreateOne_Req = ApiRequestBase<{
    payload: AcmeDnsProvider_CreateOne_Payload;
}>;
export type AcmeDnsProvider_CreateOne_Res = ApiResponseBase<{ id: string }>;

export type AcmeDnsProvider_UpdateOne_Req = ApiRequestBase<{
    id: string;
    payload: AcmeDnsProvider_UpdateOne_Payload;
}>;
export type AcmeDnsProvider_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type AcmeDnsProvider_UpdateStatus_Req = ApiRequestBase<{
    id: string;
    payload: AcmeDnsProvider_UpdateStatus_Payload;
}>;
export type AcmeDnsProvider_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type AcmeDnsProvider_DeleteOne_Req = ApiRequestBase<{
    id: string;
}>;
export type AcmeDnsProvider_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;

export type AcmeDnsProvider_TestAccess_Req = ApiRequestBase<{
    payload: AcmeDnsProvider_TestAccess_Payload;
}>;
export type AcmeDnsProvider_TestAccess_Res = ApiResponseBase<{ type: "success" }>;
