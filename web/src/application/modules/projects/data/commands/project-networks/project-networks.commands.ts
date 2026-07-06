import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectNetworksApi } from "~/projects/api/hooks";
import type {
    ProjectNetworks_CreateOne_Req,
    ProjectNetworks_CreateOne_Res,
    ProjectNetworks_DeleteOne_Req,
    ProjectNetworks_DeleteOne_Res,
    ProjectNetworks_UpdateOne_Req,
    ProjectNetworks_UpdateOne_Res,
    ProjectNetworks_UpdateStatus_Req,
    ProjectNetworks_UpdateStatus_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type CreateOneReq = ProjectNetworks_CreateOne_Req["data"];
type CreateOneRes = ProjectNetworks_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useProjectNetworksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.networks.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

type UpdateOneReq = ProjectNetworks_UpdateOne_Req["data"];
type UpdateOneRes = ProjectNetworks_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useProjectNetworksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.networks.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.networks.$.find-one-by-id"],
                    {
                        projectID: request.projectID,
                        networkID: request.networkID,
                    },
                ],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type UpdateStatusReq = ProjectNetworks_UpdateStatus_Req["data"];
type UpdateStatusRes = ProjectNetworks_UpdateStatus_Res;
type UpdateStatusOptions = Omit<UseMutationOptions<UpdateStatusRes, Error, UpdateStatusReq>, "mutationFn">;

function useUpdateStatus({ onSuccess, ...options }: UpdateStatusOptions = {}) {
    const { mutations } = useProjectNetworksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateStatus,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.networks.$.find-many-paginated"]],
            });
            void queryClient.invalidateQueries({
                queryKey: [
                    QK["projects.networks.$.find-one-by-id"],
                    {
                        projectID: request.projectID,
                        networkID: request.networkID,
                    },
                ],
            });

            onSuccess?.(response, request, ...rest);
        },
        ...options,
    });
}

type DeleteOneReq = ProjectNetworks_DeleteOne_Req["data"];
type DeleteOneRes = ProjectNetworks_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useProjectNetworksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["projects.networks.$.find-many-paginated"]],
            });
            queryClient.removeQueries({
                queryKey: [
                    QK["projects.networks.$.find-one-by-id"],
                    {
                        projectID: request.projectID,
                        networkID: request.networkID,
                    },
                ],
                exact: true,
            });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

export const ProjectNetworksCommands = Object.freeze({
    useCreateOne,
    useUpdateOne,
    useUpdateStatus,
    useDeleteOne,
});
