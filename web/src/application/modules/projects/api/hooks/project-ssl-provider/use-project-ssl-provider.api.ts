import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectSslProvider_CreateOne_Req,
    ProjectSslProvider_DeleteOne_Req,
    ProjectSslProvider_FindManyPaginated_Req,
    ProjectSslProvider_FindOneById_Req,
    ProjectSslProvider_UpdateOne_Req,
    ProjectSslProvider_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectSslProviderApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectSslProvider_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.sslProvider.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project SSL provider settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectSslProvider_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.sslProvider.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project SSL provider setting",
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
                createOne: async (data: ProjectSslProvider_CreateOne_Req["data"]) => {
                    const result = await api.projects.sslProvider.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project SSL provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectSslProvider_UpdateOne_Req["data"]) => {
                    const result = await api.projects.sslProvider.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project SSL provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectSslProvider_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.sslProvider.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project SSL provider status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectSslProvider_DeleteOne_Req["data"]) => {
                    const result = await api.projects.sslProvider.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project SSL provider setting",
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

export const useProjectSslProviderApi = createHook();
