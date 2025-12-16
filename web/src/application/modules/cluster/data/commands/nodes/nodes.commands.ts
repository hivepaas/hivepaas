import { type UseMutationOptions, useMutation, useQueryClient } from "@tanstack/react-query";
import { useNodesApi } from "~/cluster/api/hooks";
import type {
    Nodes_CreateOne_Req,
    Nodes_CreateOne_Res,
    Nodes_DeleteOne_Req,
    Nodes_DeleteOne_Res,
    Nodes_GetJoinNode_Req,
    Nodes_GetJoinNode_Res,
    Nodes_UpdateOne_Req,
    Nodes_UpdateOne_Res,
} from "~/cluster/api/services";
import { QK } from "~/cluster/data/constants";

/**
 * Delete a node command
 */
type DeleteOneReq = Nodes_DeleteOne_Req["data"];
type DeleteOneRes = Nodes_DeleteOne_Res;
type DeleteOneOptions = Omit<UseMutationOptions<DeleteOneRes, Error, DeleteOneReq>, "mutationFn">;

function useDeleteOne({ onSuccess, ...options }: DeleteOneOptions = {}) {
    const { mutations } = useNodesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.deleteOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["nodes.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

/**
 * Update a node command
 */
type UpdateOneReq = Nodes_UpdateOne_Req["data"];
type UpdateOneRes = Nodes_UpdateOne_Res;
type UpdateOneOptions = Omit<UseMutationOptions<UpdateOneRes, Error, UpdateOneReq>, "mutationFn">;

function useUpdateOne({ onSuccess, ...options }: UpdateOneOptions = {}) {
    const { mutations } = useNodesApi();

    const queryClient = useQueryClient();

    return useMutation({
        mutationFn: mutations.updateOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["nodes.$.find-many-paginated"]],
            });

            void queryClient.invalidateQueries({
                queryKey: [QK["nodes.$.find-one-by-id"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

/**
 * Create a node command
 */
type CreateOneReq = Nodes_CreateOne_Req["data"];
type CreateOneRes = Nodes_CreateOne_Res;
type CreateOneOptions = Omit<UseMutationOptions<CreateOneRes, Error, CreateOneReq>, "mutationFn">;

function useCreateOne({ onSuccess, ...options }: CreateOneOptions = {}) {
    const { mutations } = useNodesApi();

    const queryClient = useQueryClient();
    return useMutation({
        mutationFn: mutations.createOne,
        onSuccess: (response, ...rest) => {
            void queryClient.invalidateQueries({
                queryKey: [QK["nodes.$.find-many-paginated"]],
            });

            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

/**
 * Get join node command
 */
type GetJoinNodeReq = Nodes_GetJoinNode_Req["data"];
type GetJoinNodeRes = Nodes_GetJoinNode_Res;
type GetJoinNodeOptions = Omit<UseMutationOptions<GetJoinNodeRes, Error, GetJoinNodeReq>, "mutationFn">;

function useGetJoinNode({ onSuccess, ...options }: GetJoinNodeOptions = {}) {
    const { queries } = useNodesApi();

    return useMutation({
        mutationFn: (data: GetJoinNodeReq) => queries.getJoinNode(data),
        onSuccess: (response, ...rest) => {
            if (onSuccess) {
                onSuccess(response, ...rest);
            }
        },

        ...options,
    });
}

export const NodesCommands = Object.freeze({
    useDeleteOne,
    useUpdateOne,
    useCreateOne,
    useGetJoinNode,
});
