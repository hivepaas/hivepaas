import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingGitCredential } from "~/settings/domain";

import type { ApiRequestBase, ApiResponsePaginated } from "@infrastructure/api";

export type GitCredentials_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type GitCredentials_FindManyPaginated_Res = ApiResponsePaginated<SettingGitCredential>;
