import type { PaginationState } from "@infrastructure/data";

import type { AppPublic } from "@application/shared/entities";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many public apps base
 */
export type Public_Apps_FindManyBase_Req = ApiRequestBase<{
    projectID: string;
    search?: string;
    pagination?: PaginationState;
}>;

export type Public_Apps_FindManyBase_Res = ApiResponsePaginated<AppPublic>;
