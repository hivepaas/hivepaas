import { type AppHttpSettings, type AppHttpSettingsUpdatePayload } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppHttpSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppHttpSettings_FindOne_Res = ApiResponseBase<AppHttpSettings>;

export type AppHttpSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppHttpSettingsUpdatePayload;
}>;
export type AppHttpSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
