import { type AppStorageSettings } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppStorageSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppStorageSettings_FindOne_Res = ApiResponseBase<AppStorageSettings>;

export type AppStorageSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppStorageSettings;
}>;
export type AppStorageSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
