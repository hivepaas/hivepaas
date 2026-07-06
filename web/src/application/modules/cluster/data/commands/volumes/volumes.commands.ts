import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useClusterVolumesApi } from "~/cluster/api/hooks";
import type {
    ClusterVolumes_CreateOne_Req,
    ClusterVolumes_CreateOne_Res,
    ClusterVolumes_DeleteOne_Req,
    ClusterVolumes_DeleteOne_Res,
    ClusterVolumes_SyncFromDocker_Req,
    ClusterVolumes_SyncFromDocker_Res,
    ClusterVolumes_UpdateOne_Req,
    ClusterVolumes_UpdateOne_Res,
    ClusterVolumes_UpdateStatus_Req,
    ClusterVolumes_UpdateStatus_Res,
} from "~/cluster/api/services";
import { QK } from "~/cluster/data/constants";

type CreateOneReq = ClusterVolumes_CreateOne_Req["data"];
type CreateOneRes = ClusterVolumes_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.list"]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ClusterVolumes_UpdateOne_Req["data"];
type UpdateOneRes = ClusterVolumes_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.list"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.find-one-by-id"], { volumeID: request.volumeID }],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ClusterVolumes_UpdateStatus_Req["data"];
type UpdateStatusRes = ClusterVolumes_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.list"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.find-one-by-id"], { volumeID: request.volumeID }],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ClusterVolumes_DeleteOne_Req["data"];
type DeleteOneRes = ClusterVolumes_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.list"]],
            });
            queryClient.removeQueries({
                queryKey: [QK["volumes.$.find-one-by-id"], { volumeID: request.volumeID }],
                exact: true,
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type SyncFromDockerReq = ClusterVolumes_SyncFromDocker_Req["data"];
type SyncFromDockerRes = ClusterVolumes_SyncFromDocker_Res;
type SyncFromDockerOptions = Omit<UseMutationOptions<SyncFromDockerRes, Error, SyncFromDockerReq>, "mutationFn">;

function useSyncFromDocker({ onSuccess, ...options }: SyncFromDockerOptions = {}) {
    const { mutations } = useClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.syncFromDocker,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [QK["volumes.$.list"]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const ClusterVolumesCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
    useSyncFromDocker,
});
