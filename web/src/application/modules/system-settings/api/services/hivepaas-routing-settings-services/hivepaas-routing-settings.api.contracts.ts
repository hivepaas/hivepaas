import type {
    HivePaaSRoutingSettings,
    HivePaaSRoutingSettingsUpdatePayload,
    SettingsPendingChange,
} from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSRoutingSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSRoutingSettings_FindOne_Res = ApiResponseBase<HivePaaSRoutingSettings>;

export type HivePaaSRoutingSettings_UpdateOne_Req = ApiRequestBase<{
    payload: HivePaaSRoutingSettingsUpdatePayload;
}>;
// The update answers with the trial it just started, so the caller knows what to
// confirm and by when. Null only if the server did not put the change on trial.
export type HivePaaSRoutingSettings_UpdateOne_Res = ApiResponseBase<{
    pendingChange: SettingsPendingChange | null;
}>;

export type HivePaaSRoutingSettings_ConfirmChange_Req = ApiRequestBase<{ changeId: string }>;
export type HivePaaSRoutingSettings_ConfirmChange_Res = ApiResponseBase<{ type: "success" }>;

export type HivePaaSRoutingSettings_RevertChange_Req = ApiRequestBase<{ changeId: string }>;
export type HivePaaSRoutingSettings_RevertChange_Res = ApiResponseBase<{
    reverted: boolean;
    reason: string | null;
}>;

// Aliases
export type HivePaaSHttpSettings_FindOne_Req = HivePaaSRoutingSettings_FindOne_Req;
export type HivePaaSHttpSettings_FindOne_Res = HivePaaSRoutingSettings_FindOne_Res;
export type HivePaaSHttpSettings_UpdateOne_Req = HivePaaSRoutingSettings_UpdateOne_Req;
export type HivePaaSHttpSettings_UpdateOne_Res = HivePaaSRoutingSettings_UpdateOne_Res;
