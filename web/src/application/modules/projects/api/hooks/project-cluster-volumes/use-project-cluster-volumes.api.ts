import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectClusterVolumes_CreateOne_Req,
    ProjectClusterVolumes_DeleteOne_Req,
    ProjectClusterVolumes_FindManyPaginated_Req,
    ProjectClusterVolumes_FindOneById_Req,
    ProjectClusterVolumes_UpdateOne_Req,
    ProjectClusterVolumes_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectClusterVolumesApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectClusterVolumes_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.clusterVolumes.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project cluster volumes",
                                error,
                            });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectClusterVolumes_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.clusterVolumes.$.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project cluster volume",
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
                createOne: async (data: ProjectClusterVolumes_CreateOne_Req["data"]) => {
                    const result = await api.projects.clusterVolumes.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project cluster volume",
                                error,
                            });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectClusterVolumes_UpdateOne_Req["data"]) => {
                    const result = await api.projects.clusterVolumes.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project cluster volume",
                                error,
                            });
                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectClusterVolumes_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.clusterVolumes.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project cluster volume status",
                                error,
                            });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectClusterVolumes_DeleteOne_Req["data"]) => {
                    const result = await api.projects.clusterVolumes.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project cluster volume",
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

export const useProjectClusterVolumesApi = createHook();
