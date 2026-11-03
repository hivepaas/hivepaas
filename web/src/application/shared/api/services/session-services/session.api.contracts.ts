import { type Profile } from "@application/shared/entities";

import { type ApiResponseBase } from "@infrastructure/api";

/**
 * Get profile
 */
/** timezone is the installation's, a zone name such as America/New_York: what a schedule's hours are read in. */
export type Session_GetProfile_Res = ApiResponseBase<Profile & { nextStep?: string; timezone: string }>;

/**
 * Logout
 */
export type Session_Logout_Res = ApiResponseBase<{
    type: "success";
}>;
