import { type AppResourceSettings } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppResourceSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppResourceSettings_FindOne_Res = ApiResponseBase<AppResourceSettings>;

export type AppResourceSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppResourceSettings;
}>;
export type AppResourceSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
