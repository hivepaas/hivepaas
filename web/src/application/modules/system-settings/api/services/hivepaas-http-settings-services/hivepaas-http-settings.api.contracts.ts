import type { HivePaaSHttpSettings, HivePaaSHttpSettingsUpdatePayload } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSHttpSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSHttpSettings_FindOne_Res = ApiResponseBase<HivePaaSHttpSettings>;

export type HivePaaSHttpSettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSHttpSettingsUpdatePayload;
}>;
export type HivePaaSHttpSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
