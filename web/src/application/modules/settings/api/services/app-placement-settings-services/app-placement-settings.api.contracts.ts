import type { AppPlacementSettings } from "~/settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type AppPlacementSettings_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type AppPlacementSettings_FindOne_Res = ApiResponseBase<AppPlacementSettings>;

export type AppPlacementSettings_UpdateOne_Payload = {
    updateVer: number;
    inheritable?: boolean;
    default?: boolean;
    excludeManagerNodes: boolean;
    excludeBuildNodes: boolean;
};

export type AppPlacementSettings_UpdateOne_Req = ApiRequestBase<{
    payload: AppPlacementSettings_UpdateOne_Payload;
}>;
export type AppPlacementSettings_UpdateOne_Res = ApiResponseBase<{ type: "success" }>;
