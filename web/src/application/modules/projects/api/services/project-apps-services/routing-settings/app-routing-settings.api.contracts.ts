import { type AppRoutingSettings, type AppRoutingSettingsUpdatePayload } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppRoutingSettings_FindOne_Req = ApiRequestBase<{ projectID: string; env: string; appID: string }>;
export type AppRoutingSettings_FindOne_Res = ApiResponseBase<AppRoutingSettings>;

export type AppRoutingSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: AppRoutingSettingsUpdatePayload;
}>;
export type AppRoutingSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;

// Aliases
export type AppHttpSettings_FindOne_Req = AppRoutingSettings_FindOne_Req;
export type AppHttpSettings_FindOne_Res = AppRoutingSettings_FindOne_Res;
export type AppHttpSettings_UpdateOne_Req = AppRoutingSettings_UpdateOne_Req;
export type AppHttpSettings_UpdateOne_Res = AppRoutingSettings_UpdateOne_Res;
