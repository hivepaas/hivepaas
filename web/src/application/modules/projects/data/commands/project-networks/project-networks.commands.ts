import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useProjectNetworksApi } from "~/projects/api/hooks";
import type {
    ProjectNetworks_CreateOne_Req,
    ProjectNetworks_CreateOne_Res,
    ProjectNetworks_DeleteOne_Req,
    ProjectNetworks_DeleteOne_Res,
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
    useDeleteOne,
});
