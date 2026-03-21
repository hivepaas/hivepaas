import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectRegistryAuth_CreateOne_Req,
    ProjectRegistryAuth_DeleteOne_Req,
    ProjectRegistryAuth_FindManyPaginated_Req,
    ProjectRegistryAuth_FindOneById_Req,
    ProjectRegistryAuth_UpdateMeta_Req,
    ProjectRegistryAuth_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectRegistryAuthApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectRegistryAuth_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.registryAuth.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project registry auth settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectRegistryAuth_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.registryAuth.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project registry auth",
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
                createOne: async (data: ProjectRegistryAuth_CreateOne_Req["data"]) => {
                    const result = await api.projects.registryAuth.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project registry auth",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectRegistryAuth_UpdateOne_Req["data"]) => {
                    const result = await api.projects.registryAuth.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project registry auth",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateMeta: async (data: ProjectRegistryAuth_UpdateMeta_Req["data"]) => {
                    const result = await api.projects.registryAuth.$.updateMeta({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project registry auth meta",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectRegistryAuth_DeleteOne_Req["data"]) => {
                    const result = await api.projects.registryAuth.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project registry auth",
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

export const useProjectRegistryAuthApi = createHook();
