import type { McpSettings } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type McpSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type McpSettings_FindOne_Res = ApiResponseBase<McpSettings>;

export type McpSettings_UpdateOne_Req = ApiRequestBase<{
    payload: {
        enabled: boolean;
        allowWrite: boolean;
        updateVer: number;
    };
}>;
export type McpSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
