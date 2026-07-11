import type { SystemSslRenewalSettings } from "~/system-settings/domain";

import type { ESettingStatus } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type SystemSslRenewal_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type SystemSslRenewal_FindOne_Res = ApiResponseBase<SystemSslRenewalSettings>;

export type SystemSslRenewal_UpdateOne_Payload = {
    updateVer: number;
    status: ESettingStatus;
    schedule: {
        interval: string;
        cronExpr: string;
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

export type SystemSslRenewal_UpdateOne_Req = ApiRequestBase<{
    payload: SystemSslRenewal_UpdateOne_Payload;
}>;
export type SystemSslRenewal_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

export type SystemSslRenewal_Execute_Req = ApiRequestBase<{
    targetSSLs: { id: string }[];
}>;
export type SystemSslRenewal_Execute_Res = ApiResponseBase<{
    task: {
        id: string;
    };
}>;
