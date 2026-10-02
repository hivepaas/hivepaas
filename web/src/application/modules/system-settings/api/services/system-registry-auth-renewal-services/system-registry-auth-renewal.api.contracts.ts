import type { SystemRegistryAuthRenewalSettings } from "~/system-settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SystemRegistryAuthRenewal_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type SystemRegistryAuthRenewal_FindOne_Res = ApiResponseBase<SystemRegistryAuthRenewalSettings>;

export type SystemRegistryAuthRenewal_UpdateOne_Payload = {
    updateVer: number;
    status: ESettingStatus;
    schedule: {
        interval: string;
        initialTime?: Date;
    };
    notification: {
        success: {
            id: string;
        };
        successUseDefault: boolean;
        failure: {
            id: string;
        };
        failureUseDefault: boolean;
    };
};

export type SystemRegistryAuthRenewal_UpdateOne_Req = ApiRequestBase<{
    payload: SystemRegistryAuthRenewal_UpdateOne_Payload;
}>;
export type SystemRegistryAuthRenewal_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SystemRegistryAuthRenewal_Execute_Req = ApiRequestBase<{
    targetAuths: { id: string }[];
}>;
export type SystemRegistryAuthRenewal_Execute_Res = ApiResponseBase<{
    task: {
        id: string;
    };
}>;
