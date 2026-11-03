import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    ClusterNetwork,
    ClusterNetworkCreatePayload,
    ClusterNetworkUpdatePayload,
    ClusterNetworkUpdateStatusPayload,
} from "~/cluster/domain";

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
    payload: ClusterNetworkCreatePayload;
}>;

export type ClusterNetworks_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type ClusterNetworks_UpdateOne_Req = ApiRequestBase<{
    networkID: string;
    payload: ClusterNetworkUpdatePayload;
}>;

export type ClusterNetworks_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type ClusterNetworks_UpdateStatus_Req = ApiRequestBase<{
    networkID: string;
    payload: ClusterNetworkUpdateStatusPayload;
}>;

export type ClusterNetworks_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

export type ClusterNetworks_DeleteOne_Req = ApiRequestBase<{
    networkID: string;
}>;

export type ClusterNetworks_DeleteOne_Res = ApiResponseBase<{
    networkID: string;
}>;

export type ClusterNetworks_SyncFromDocker_Req = ApiRequestBase<Record<string, never>>;

export type ClusterNetworks_SyncFromDocker_Res = ApiResponseBase<{
    type: "success";
}>;
