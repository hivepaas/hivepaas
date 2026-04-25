import type { PaginationState, SortingState } from "@infrastructure/data";
import type { ProjectNotification } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectNotification_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectNotification_FindManyPaginated_Res = ApiResponsePaginated<ProjectNotification>;

export type ProjectNotification_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    id: string;
}>;

export type ProjectNotification_FindOneById_Res = ApiResponseBase<ProjectNotification>;
