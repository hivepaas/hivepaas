import { type Profile } from "@application/shared/entities";

import { type ApiResponseBase } from "@infrastructure/api";

/**
 * Get profile
 */
export type Session_GetProfile_Res = ApiResponseBase<Profile>;

/**
 * Logout
 */
export type Session_Logout_Res = ApiResponseBase<{
    type: "success";
}>;
