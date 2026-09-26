import type { DashboardCert } from "~/home/domain";

import type { ApiRequestBase, ApiResponseBase } from "@infrastructure/api";

export type GetStarted_GetDashboardCert_Req = ApiRequestBase<Record<string, never>>;
export type GetStarted_GetDashboardCert_Res = ApiResponseBase<DashboardCert>;

export type GetStarted_RequestDashboardCert_Req = ApiRequestBase<Record<string, never>>;
export type GetStarted_RequestDashboardCert_Res = ApiResponseBase<DashboardCert>;

export type GetStarted_Dismiss_Req = ApiRequestBase<Record<string, never>>;
export type GetStarted_Dismiss_Res = ApiResponseBase<{ type: "success" }>;
