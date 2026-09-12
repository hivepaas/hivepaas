import type { HivePaaSLoggingSettings, HivePaaSLoggingStatus } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSLoggingSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSLoggingSettings_FindOne_Res = ApiResponseBase<{
    settings: HivePaaSLoggingSettings;
    status: HivePaaSLoggingStatus;
}>;

export type HivePaaSLoggingSettings_UpdateOne_Req = ApiRequestBase<{ payload: HivePaaSLoggingSettings }>;
export type HivePaaSLoggingSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
