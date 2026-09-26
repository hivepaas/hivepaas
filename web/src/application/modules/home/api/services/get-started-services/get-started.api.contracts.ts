import type { SetupChecklistItem } from "@application/shared/entities";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type GetStarted_RequestDashboardCert_Req = ApiRequestBase<Record<string, never>>;
export type GetStarted_RequestDashboardCert_Res = ApiResponseBase<SetupChecklistItem>;

export type GetStarted_Dismiss_Req = ApiRequestBase<Record<string, never>>;
export type GetStarted_Dismiss_Res = ApiResponseBase<{ type: "success" }>;
