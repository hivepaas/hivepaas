import type { PaginationState, SortingState } from "@infrastructure/data";

import type { SshKeyPublicEntity } from "@application/shared/entities";

import { type ApiRequestBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many SSH keys paginated
 */
export type SshKeys_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    status?: string;
    search?: string;
}>;

export type SshKeys_FindManyPaginated_Res = ApiResponsePaginated<SshKeyPublicEntity>;
