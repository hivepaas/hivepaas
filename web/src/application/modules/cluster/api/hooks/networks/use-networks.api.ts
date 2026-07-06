import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { NodesApiContext } from "~/cluster/api/api-context";
import type {
    ClusterNetworks_CreateOne_Req,
    ClusterNetworks_DeleteOne_Req,
    ClusterNetworks_FindManyPaginated_Req,
    ClusterNetworks_FindOneById_Req,
    ClusterNetworks_UpdateOne_Req,
    ClusterNetworks_UpdateStatus_Req,
} from "~/cluster/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useClusterNetworksApi() {
        const { api } = use(NodesApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ClusterNetworks_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.networks.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get cluster networks",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ClusterNetworks_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.networks.$.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get cluster network",
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
                createOne: async (data: ClusterNetworks_CreateOne_Req["data"]) => {
                    const result = await api.networks.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create cluster network",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ClusterNetworks_UpdateOne_Req["data"]) => {
                    const result = await api.networks.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update cluster network",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ClusterNetworks_UpdateStatus_Req["data"]) => {
                    const result = await api.networks.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update cluster network status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ClusterNetworks_DeleteOne_Req["data"]) => {
                    const result = await api.networks.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete cluster network",
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

export const useClusterNetworksApi = createHook();
