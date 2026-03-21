import type { ProjectGitRepo } from "~/projects/domain";

import type { PaginationState, SortingState } from "@infrastructure/data";
import type { SettingGitCredential } from "~/settings/domain";

import type { ApiRequestBase, ApiResponsePaginated } from "@infrastructure/api";

export type { ProjectGitRepo };

export type ProjectGitCredentials_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectGitCredentials_FindManyPaginated_Res = ApiResponsePaginated<SettingGitCredential>;

export type ProjectGitCredentials_FindManyRepos_Req = ApiRequestBase<{
    projectID: string;
    itemID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectGitCredentials_FindManyRepos_Res = ApiResponsePaginated<ProjectGitRepo>;
