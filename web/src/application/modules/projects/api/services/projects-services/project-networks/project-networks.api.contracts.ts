import { type PaginationState, type SortingState } from "@infrastructure/data";
import type { EClusterNetworkDriver } from "~/cluster/module-shared/enums";
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
    payload: {
        name: string;
        driver: EClusterNetworkDriver;
        enableIPv4: boolean;
        enableIPv6: boolean;
        internal: boolean;
        attachable: boolean;
        ingress: boolean;
        labels: Record<string, string>;
        options: Record<string, string>;
    };
}>;

export type ProjectNetworks_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type ProjectNetworks_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    networkID: string;
}>;

export type ProjectNetworks_DeleteOne_Res = ApiResponseBase<{
    networkID: string;
}>;
