import { type PaginationState, type SortingState } from "@infrastructure/data";
import type {
    ClusterNetworkCreatePayload,
    ClusterNetworkUpdatePayload,
    ClusterNetworkUpdateStatusPayload,
} from "~/cluster/domain";
import { type ProjectNetworkEntity } from "~/projects/domain";

import { type ApiRequestBase, type ApiResponseBase, type ApiResponsePaginated } from "@infrastructure/api";

/**
 * Find many project networks paginated
 */
export type ProjectNetworks_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectNetworks_FindManyPaginated_Res = ApiResponsePaginated<ProjectNetworkEntity>;

export type ProjectNetworks_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    networkID: string;
}>;

export type ProjectNetworks_FindOneById_Res = ApiResponseBase<ProjectNetworkEntity>;

export type ProjectNetworks_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: Omit<ClusterNetworkCreatePayload, "availableInProjects"> & {
        availableInProjects?: false;
    };
}>;

export type ProjectNetworks_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type ProjectNetworks_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    networkID: string;
    payload: Omit<ClusterNetworkUpdatePayload, "availableInProjects"> & {
        availableInProjects?: false;
    };
}>;

export type ProjectNetworks_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type ProjectNetworks_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    networkID: string;
    payload: Omit<ClusterNetworkUpdateStatusPayload, "availableInProjects"> & {
        availableInProjects?: false;
    };
}>;

export type ProjectNetworks_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

export type ProjectNetworks_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    networkID: string;
}>;

export type ProjectNetworks_DeleteOne_Res = ApiResponseBase<{
    networkID: string;
}>;
