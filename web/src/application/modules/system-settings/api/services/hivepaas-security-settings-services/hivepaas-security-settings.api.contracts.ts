import type { HivePaaSSecuritySettings, HivePaaSSecuritySettingsUpdatePayload } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSSecuritySettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSSecuritySettings_FindOne_Res = ApiResponseBase<HivePaaSSecuritySettings>;

export type HivePaaSSecuritySettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSSecuritySettingsUpdatePayload;
}>;
export type HivePaaSSecuritySettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
