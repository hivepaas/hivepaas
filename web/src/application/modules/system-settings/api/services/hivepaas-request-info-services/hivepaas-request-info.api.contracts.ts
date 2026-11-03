import type { HivePaaSRequestInfo } from "~/system-settings/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type HivePaaSRequestInfo_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type HivePaaSRequestInfo_FindOne_Res = ApiResponseBase<HivePaaSRequestInfo>;
