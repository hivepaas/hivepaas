import type { PaginationState, SortingState } from "@infrastructure/data";
import type {
    ClusterVolume,
    ClusterVolumeBasePayload,
    ClusterVolumeUpdatePayload,
    ClusterVolumeUpdateStatusPayload,
} from "~/cluster/domain";

import type { ApiRequestBase, ApiResponseBase, ApiResponsePaginated } from "@infrastructure/api";

export type ProjectClusterVolumes_FindManyPaginated_Req = ApiRequestBase<{
    projectID: string;
    pagination?: PaginationState;
    sorting?: SortingState;
    search?: string;
}>;

export type ProjectClusterVolumes_FindManyPaginated_Res = ApiResponsePaginated<ClusterVolume>;

export type ProjectClusterVolumes_FindOneById_Req = ApiRequestBase<{
    projectID: string;
    volumeID: string;
}>;

export type ProjectClusterVolumes_FindOneById_Res = ApiResponseBase<ClusterVolume>;

export type ProjectClusterVolumes_CreateOne_Req = ApiRequestBase<{
    projectID: string;
    payload: ClusterVolumeBasePayload & {
        default: boolean;
        availableInProjects?: false;
    };
}>;

export type ProjectClusterVolumes_CreateOne_Res = ApiResponseBase<{
    id: string;
}>;

export type ProjectClusterVolumes_UpdateOne_Req = ApiRequestBase<{
    projectID: string;
    volumeID: string;
    payload: Omit<ClusterVolumeUpdatePayload, "availableInProjects"> & {
        availableInProjects?: false;
    };
}>;

export type ProjectClusterVolumes_UpdateOne_Res = ApiResponseBase<{
    type: "success";
}>;

export type ProjectClusterVolumes_UpdateStatus_Req = ApiRequestBase<{
    projectID: string;
    volumeID: string;
    payload: Omit<ClusterVolumeUpdateStatusPayload, "availableInProjects"> & {
        availableInProjects?: false;
    };
}>;

export type ProjectClusterVolumes_UpdateStatus_Res = ApiResponseBase<{
    type: "success";
}>;

export type ProjectClusterVolumes_DeleteOne_Req = ApiRequestBase<{
    projectID: string;
    volumeID: string;
}>;

export type ProjectClusterVolumes_DeleteOne_Res = ApiResponseBase<{
    type: "success";
}>;
