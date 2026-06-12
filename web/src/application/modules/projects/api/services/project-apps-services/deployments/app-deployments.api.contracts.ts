import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AppDeployment } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AppDeployments_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppDeployments_FindManyPaginated_Res = ApiResponsePaginated<AppDeployment>;

export type AppDeployments_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    deploymentID: string;
}>;

export type AppDeployments_FindOneById_Res = ApiResponseBase<AppDeployment>;

export type AppDeployments_Cancel_Req = ApiRequestBase<{
    projectID: string;
    appID: string;
    deploymentID: string;
}>;

export type AppDeployments_Cancel_Res = ApiResponseBase<{
    canceled: boolean;
}>;
