import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectClusterVolumesApi } from "~/projects/api/hooks";
import type {
    ProjectClusterVolumes_CreateOne_Req,
    ProjectClusterVolumes_CreateOne_Res,
    ProjectClusterVolumes_DeleteOne_Req,
    ProjectClusterVolumes_DeleteOne_Res,
    ProjectClusterVolumes_UpdateOne_Req,
    ProjectClusterVolumes_UpdateOne_Res,
    ProjectClusterVolumes_UpdateStatus_Req,
    ProjectClusterVolumes_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectClusterVolumes_CreateOne_Req["data"];
type CreateOneRes = ProjectClusterVolumes_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.cluster-volumes.$.find-many-paginated"]],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type UpdateOneReq = ProjectClusterVolumes_UpdateOne_Req["data"];
type UpdateOneRes = ProjectClusterVolumes_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.cluster-volumes.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.cluster-volumes.$.find-one-by-id"],
                    {
                        projectID: request.projectID,
                        volumeID: request.volumeID,
                    },
                ],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectClusterVolumes_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectClusterVolumes_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.cluster-volumes.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.cluster-volumes.$.find-one-by-id"],
                    {
                        projectID: request.projectID,
                        volumeID: request.volumeID,
                    },
                ],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectClusterVolumes_DeleteOne_Req["data"];
type DeleteOneRes = ProjectClusterVolumes_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectClusterVolumesApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.cluster-volumes.$.find-many-paginated"]],
            });
            queryClient.removeQueries({
                queryKey: [
                    QK["projects.cluster-volumes.$.find-one-by-id"],
                    {
                        projectID: request.projectID,
                        volumeID: request.volumeID,
                    },
                ],
                exact: true,
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

export const ProjectClusterVolumesCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
