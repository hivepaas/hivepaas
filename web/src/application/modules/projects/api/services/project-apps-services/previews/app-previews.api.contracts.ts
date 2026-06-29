import type { PaginationState, SortingState } from "@infrastructure/data";
import type { ProjectAppDetails } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AppPreviews_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppPreviews_FindManyPaginated_Res = ApiResponsePaginated<ProjectAppDetails>;

export type AppPreviews_PrepareCreate_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
}>;

export type AppPreviews_PrepareCreate_Res = ApiResponseBase<{
    repoURL: string;
    repoCredentials: { id: string } | null;
    canListBranches: boolean;
    canListPullRequests: boolean;
}>;

export type AppPreviews_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    repoRef: string;
    customSubdomain: string;
    noStart: boolean;
}>;

export type AppPreviews_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;
