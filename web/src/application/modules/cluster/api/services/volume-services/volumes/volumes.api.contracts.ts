import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    ClusterVolume,
    ClusterVolumeCreatePayload,
    ClusterVolumeUpdatePayload,
    ClusterVolumeUpdateStatusPayload,
} from "~/cluster/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ClusterVolumes_FindManyPaginated_Req = ApiRequestBase<{
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
    type?: "cluster" | "volume";
}>;

export type ClusterVolumes_FindManyPaginated_Res = ApiResponsePaginated<ClusterVolume>;

export type ClusterVolumes_FindOneById_Req = ApiRequestBase<{
    volumeID: string;
}>;

export type ClusterVolumes_FindOneById_Res = ApiResponseBase<ClusterVolume>;

export type ClusterVolumes_CreateOne_Req = ApiRequestBase<{
    payload: ClusterVolumeCreatePayload;
}>;

export type ClusterVolumes_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type ClusterVolumes_UpdateOne_Req = ApiRequestBase<{
    volumeID: string;
    payload: ClusterVolumeUpdatePayload;
}>;

export type ClusterVolumes_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type ClusterVolumes_UpdateStatus_Req = ApiRequestBase<{
    volumeID: string;
    payload: ClusterVolumeUpdateStatusPayload;
}>;

export type ClusterVolumes_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

export type ClusterVolumes_DeleteOne_Req = ApiRequestBase<{
    volumeID: string;
}>;

export type ClusterVolumes_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type ClusterVolumes_SyncFromDocker_Req = ApiRequestBase<Record<string, never>>;

export type ClusterVolumes_SyncFromDocker_Res = ApiResponseBase<{
    type: "success";
}>;

export type ClusterVolumes_List_Req = ClusterVolumes_FindManyPaginated_Req;
export type ClusterVolumes_List_Res = ClusterVolumes_FindManyPaginated_Res;
