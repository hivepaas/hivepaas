import type { HivePaaSReleaseInfo, HivePaaSUpdatePlan } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSUpdates_FindReleaseInfo_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSUpdates_FindReleaseInfo_Res = ApiResponseBase<HivePaaSReleaseInfo>;

export type HivePaaSUpdates_FindPlan_Req = ApiRequestBase<{ targetVersion: string }>;
export type HivePaaSUpdates_FindPlan_Res = ApiResponseBase<HivePaaSUpdatePlan>;

export type HivePaaSUpdates_Update_Req = ApiRequestBase<{
    targetVersion: string;
    /** Updates without dumping the database first. A failed migration then cannot be undone. */
    skipBackup: boolean;
}>;
export type HivePaaSUpdates_Update_Res = ApiResponseBase<{ type: "success" }>;
