import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { NodesApiContext } from "~/cluster/api/api-context";
import type {
    Nodes_CreateOne_Req,
    Nodes_DeleteOne_Req,
    Nodes_FindManyPaginated_Req,
    Nodes_FindOneById_Req,
    Nodes_GetJoinNode_Req,
    Nodes_UpdateOne_Req,
} from "~/cluster/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useNodesApi() {
        const { api } = use(NodesApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many nodes paginated
                 */
                findManyPaginated: async (data: Nodes_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.nodes.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get nodes",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Find one node by id
                 */
                findOneById: async (data: Nodes_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.nodes.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get node",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Get join node command
                 */
                getJoinNode: async (data: Nodes_GetJoinNode_Req["data"], signal?: AbortSignal) => {
                    const result = await api.nodes.$.getJoinNode(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get join node command",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                /**
                 * Delete a node
                 */
                deleteOne: async (data: Nodes_DeleteOne_Req["data"]) => {
                    const result = await api.nodes.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete node",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update a node
                 */
                updateOne: async (data: Nodes_UpdateOne_Req["data"]) => {
                    const result = await api.nodes.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update node",
                                error,
                            });

                            throw error;
                        },
                    });
                },

                /**
                 * Create a node
                 */
                createOne: async (data: Nodes_CreateOne_Req["data"]) => {
                    const result = await api.nodes.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create node",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return {
            queries,
            mutations,
        };
    };
}

export const useNodesApi = createHook();
