import type { SettingDomainSettings } from "~/settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type DomainSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type DomainSettings_FindOne_Res = ApiResponseBase<SettingDomainSettings>;

export type DomainSettings_UpdateOne_Payload = {
    updateVer: number;
    rootDomain: string;
    certSettings?: SettingDomainSettings["certSettings"];
};

export type DomainSettings_UpdateOne_Req = ApiRequestBase<{
    payload: DomainSettings_UpdateOne_Payload;
}>;
export type DomainSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type DomainSettings_UpdateStatus_Payload = {
    updateVer: number;
    status?: ESettingStatus;
    expireAt?: Date | null;
    availableInProjects?: boolean;
    default?: boolean;
};

export type DomainSettings_UpdateStatus_Req = ApiRequestBase<{
    payload: DomainSettings_UpdateStatus_Payload;
}>;
export type DomainSettings_UpdateStatus_Res = ApiResponseBase<{ type: "success" }>;

export type DomainSettings_DeleteOne_Req = ApiRequestBase<Record<string, never>>;
export type DomainSettings_DeleteOne_Res = ApiResponseBase<{ type: "success" }>;
