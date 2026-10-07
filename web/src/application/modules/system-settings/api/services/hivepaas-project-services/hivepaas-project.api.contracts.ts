import type { HivePaaSProject } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSProject_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSProject_FindOne_Res = ApiResponseBase<HivePaaSProject>;
