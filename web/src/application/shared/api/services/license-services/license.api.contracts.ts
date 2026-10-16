import type { License } from "@application/shared/entities";
import type { ELicenseRefreshResult } from "@application/shared/enums";

import type { ApiRequestBase, ApiResponseBase, OpenApiConstant } from "@infrastructure/api";

/**
 * The license in force
 */
export type License_FindOne_Req = ApiRequestBase<Record<string, never>>;
export type License_FindOne_Res = ApiResponseBase<License>;

/**
 * Install a license key, in place of the one stored, whatever its plan. Answers
 * with the license the key puts in force.
 */
export type License_InstallKey_Req = ApiRequestBase<{ key: string }>;
export type License_InstallKey_Res = ApiResponseBase<License>;

/**
 * Ask the license portal now for a newer key. Answers with what the portal said
 * and the license in force after it.
 */
export type License_Refresh_Req = ApiRequestBase<Record<string, never>>;
export type License_Refresh_Res = ApiResponseBase<{
    result: OpenApiConstant<ELicenseRefreshResult>;
    license: License;
}>;
