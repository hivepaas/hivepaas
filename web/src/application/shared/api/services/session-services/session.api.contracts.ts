import { type Profile, type SetupChecklist } from "@application/shared/entities";

import { type ApiResponseBase } from "@infrastructure/api";

/**
 * Get profile. `setupChecklist` is what the installation still has to do, given
 * to an admin while `nextStep` is `hivepaas/get-started`.
 */
export type Session_GetProfile_Res = ApiResponseBase<
    Profile & { nextStep?: string; setupChecklist: SetupChecklist | null }
>;

/**
 * Logout
 */
export type Session_Logout_Res = ApiResponseBase<{
    type: "success";
}>;
