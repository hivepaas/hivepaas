import { type AppNetworkSettings } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppNetworkSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppNetworkSettings_FindOne_Res = ApiResponseBase<AppNetworkSettings>;

export type AppNetworkSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppNetworkSettings;
}>;
export type AppNetworkSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
