import type { PaginationState, SortingState } from "@infrastructure/data";
import type { ClusterNetwork } from "~/cluster/domain";
import type { EClusterNetworkDriver } from "~/cluster/module-shared/enums";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ClusterNetworks_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ClusterNetworks_FindManyPaginated_Res = ApiResponsePaginated<ClusterNetwork>;

export type ClusterNetworks_FindOneById_Req = ApiRequestBase<{
    networkID: string;
}>;

export type ClusterNetworks_FindOneById_Res = ApiResponseBase<ClusterNetwork>;

export type ClusterNetworks_CreateOne_Req = ApiRequestBase<{
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
        availableInProjects: boolean;
    };
}>;

export type ClusterNetworks_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type ClusterNetworks_DeleteOne_Req = ApiRequestBase<{
    networkID: string;
}>;

export type ClusterNetworks_DeleteOne_Res = ApiResponseBase<{
    networkID: string;
}>;
