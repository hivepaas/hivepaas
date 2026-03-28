import { type AppContainerSettings } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppContainerSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppContainerSettings_FindOne_Res = ApiResponseBase<AppContainerSettings>;

export type AppContainerSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    payload: AppContainerSettings;
}>;
export type AppContainerSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
