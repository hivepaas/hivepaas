import type { PaginationState, SortingState } from "@infrastructure/data";
import type { ProjectGitRepo } from "~/projects/domain";
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

export interface ProjectGitBranch {
    name: string;
    sha: string;
    ref: string;
}

export interface ProjectGitPullRequest {
    id: string;
    number: number;
    title: string;
    state: string;
    branch: string;
    sha: string;
    ref: string;
    author: string;
    htmlURL: string;
    createdAt: Date;
    updatedAt: Date;
}

export type ProjectGitCredentials_FindManyBranches_Req = ApiRequestBase<{
    projectID: string;
    itemID: string;
    owner?: string;
    repo: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectGitCredentials_FindManyBranches_Res = ApiResponsePaginated<ProjectGitBranch>;

export type ProjectGitCredentials_FindManyPullRequests_Req = ApiRequestBase<{
    projectID: string;
    itemID: string;
    owner?: string;
    repo: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectGitCredentials_FindManyPullRequests_Res = ApiResponsePaginated<ProjectGitPullRequest>;
