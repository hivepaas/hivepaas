import type { HivePaaSRoutingSettings, HivePaaSRoutingSettingsUpdatePayload } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSRoutingSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSRoutingSettings_FindOne_Res = ApiResponseBase<HivePaaSRoutingSettings>;

export type HivePaaSRoutingSettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSRoutingSettingsUpdatePayload;
}>;
export type HivePaaSRoutingSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

// Aliases
export type HivePaaSHttpSettings_FindOne_Req = HivePaaSRoutingSettings_FindOne_Req;
export type HivePaaSHttpSettings_FindOne_Res = HivePaaSRoutingSettings_FindOne_Res;
export type HivePaaSHttpSettings_UpdateOne_Req = HivePaaSRoutingSettings_UpdateOne_Req;
export type HivePaaSHttpSettings_UpdateOne_Res = HivePaaSRoutingSettings_UpdateOne_Res;
