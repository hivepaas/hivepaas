import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useClusterNetworksApi } from "~/cluster/api/hooks";
import type {
    ClusterNetworks_CreateOne_Req,
    ClusterNetworks_CreateOne_Res,
    ClusterNetworks_DeleteOne_Req,
    ClusterNetworks_DeleteOne_Res,
} from "~/cluster/api/services";
import { QK } from "~/cluster/data/constants";

type CreateOneReq = ClusterNetworks_CreateOne_Req["data"];
type CreateOneRes = ClusterNetworks_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useClusterNetworksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["networks.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },
        ...options,
    });
}

type DeleteOneReq = ClusterNetworks_DeleteOne_Req["data"];
type DeleteOneRes = ClusterNetworks_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useClusterNetworksApi();
    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, request, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["networks.$.find-many-paginated"]],
            });
            queryClient.removeQueries({
                queryKey: [QK["networks.$.find-one-by-id"], { networkID: request.networkID }],
                exact: true,
            });

            if (onSuccess) {
                onSuccess(response, request, ...rest);
            }
        },
        ...options,
    });
}

export const ClusterNetworksCommands = Object.freeze({
    useCreateOne,
    useDeleteOne,
});
