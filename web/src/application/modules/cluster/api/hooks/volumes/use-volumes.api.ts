import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { NodesApiContext } from "~/cluster/api/api-context";
import type {
    ClusterVolumes_CreateOne_Req,
    ClusterVolumes_DeleteOne_Req,
    ClusterVolumes_FindManyPaginated_Req,
    ClusterVolumes_FindOneById_Req,
    ClusterVolumes_List_Req,
    ClusterVolumes_SyncFromDocker_Req,
    ClusterVolumes_UpdateOne_Req,
    ClusterVolumes_UpdateStatus_Req,
} from "~/cluster/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useClusterVolumesApi() {
        const { api } = use(NodesApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: ClusterVolumes_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.volumes.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get cluster volumes",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ClusterVolumes_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.volumes.$.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get cluster volume",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                list: async (data: ClusterVolumes_List_Req["data"], signal?: AbortSignal) => {
                    const result = await api.volumes.$.list({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get cluster volumes",
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
                createOne: async (data: ClusterVolumes_CreateOne_Req["data"]) => {
                    const result = await api.volumes.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create cluster volume",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ClusterVolumes_UpdateOne_Req["data"]) => {
                    const result = await api.volumes.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update cluster volume",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ClusterVolumes_UpdateStatus_Req["data"]) => {
                    const result = await api.volumes.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update cluster volume status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ClusterVolumes_DeleteOne_Req["data"]) => {
                    const result = await api.volumes.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete cluster volume",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                syncFromDocker: async (data: ClusterVolumes_SyncFromDocker_Req["data"]) => {
                    const result = await api.volumes.$.syncFromDocker({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to sync cluster volumes from Docker",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useClusterVolumesApi = createHook();
