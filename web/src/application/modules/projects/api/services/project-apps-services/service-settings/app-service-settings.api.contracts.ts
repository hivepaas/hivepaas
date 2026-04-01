import { type AppServiceSettings } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppServiceSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppServiceSettings_FindOne_Res = ApiResponseBase<AppServiceSettings>;

export type AppServiceSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppServiceSettings;
}>;
export type AppServiceSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
