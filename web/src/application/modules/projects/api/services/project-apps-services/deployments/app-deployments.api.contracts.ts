import type { PaginationState, SortingState } from "@infrastructure/data";
import type { AppActiveDeployment, AppDeployment } from "~/projects/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type AppDeployments_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type AppDeployments_FindManyPaginated_Res = ApiResponsePaginated<AppDeployment>;

export type AppDeployments_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    deploymentID: string;
}>;

export type AppDeployments_FindOneById_Res = ApiResponseBase<AppDeployment>;

export type AppDeployments_Cancel_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
    deploymentID: string;
}>;

export type AppDeployments_Cancel_Res = ApiResponseBase<{
    canceled: boolean;
}>;

export type AppDeployments_FindActive_Req = ApiRequestBase<{
    projectID: string;
    env: string;
    appID: string;
}>;

/** Null when no deployment of the app is queued or running. */
export type AppDeployments_FindActive_Res = ApiResponseBase<AppActiveDeployment | null>;
