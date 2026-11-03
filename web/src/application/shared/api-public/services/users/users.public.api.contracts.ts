import type { UserPublic } from "@application/shared/entities";
import type { EUserRole } from "@application/shared/enums";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many public users
 */
export type Public_Users_FindMany_Req = ApiRequestBase<{
    search?: string;
    role?: EUserRole;
}>;

export type Public_Users_FindMany_Res = ApiResponsePaginated<UserPublic>;
