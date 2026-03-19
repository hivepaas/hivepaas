import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

import { type AppDeploymentSettings } from "../../../../domain/apps/deployment-settings";

export type AppDeploymentSettings_FindOne_Req = ApiRequestBase<{ projectID: string; appID: string }>;
export type AppDeploymentSettings_FindOne_Res = ApiResponseBase<AppDeploymentSettings>;

export type AppDeploymentSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    updateVer: number;
    payload: AppDeploymentSettings;
}>;
export type AppDeploymentSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
