import { type AppKindSettings, type AppKindSettingsUpdatePayload } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase } from "@infrastructure/api";

export type AppKindSettings_FindOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    revealSecrets?: boolean;
}>;

export type AppKindSettings_FindOne_Res = ApiResponseBase<AppKindSettings>;

export type AppKindSettings_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    payload: AppKindSettingsUpdatePayload;
}>;

export type AppKindSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
