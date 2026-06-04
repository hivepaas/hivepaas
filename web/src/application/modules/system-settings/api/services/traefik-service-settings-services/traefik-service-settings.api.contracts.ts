import type { TraefikServiceSettings } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type TraefikServiceSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type TraefikServiceSettings_FindOne_Res = ApiResponseBase<TraefikServiceSettings>;

export type TraefikServiceSettings_UpdateOne_Payload = {
    updateVer: number;
    appSettings: {
        replicas: number;
    };
};

export type TraefikServiceSettings_UpdateOne_Req = ApiRequestBase<{
    payload: TraefikServiceSettings_UpdateOne_Payload;
}>;
export type TraefikServiceSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
