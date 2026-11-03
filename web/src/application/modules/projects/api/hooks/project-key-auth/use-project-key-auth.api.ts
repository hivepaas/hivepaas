import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectKeyAuth_CreateOne_Req,
    ProjectKeyAuth_DeleteOne_Req,
    ProjectKeyAuth_FindManyPaginated_Req,
    ProjectKeyAuth_FindOneById_Req,
    ProjectKeyAuth_UpdateOne_Req,
    ProjectKeyAuth_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectKeyAuthApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: ProjectKeyAuth_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.keyAuth.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project key auth settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectKeyAuth_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.keyAuth.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project key auth setting",
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
                createOne: async (data: ProjectKeyAuth_CreateOne_Req["data"]) => {
                    const result = await api.projects.keyAuth.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project key auth setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectKeyAuth_UpdateOne_Req["data"]) => {
                    const result = await api.projects.keyAuth.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project key auth setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectKeyAuth_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.keyAuth.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project key auth status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectKeyAuth_DeleteOne_Req["data"]) => {
                    const result = await api.projects.keyAuth.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project key auth setting",
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

export const useProjectKeyAuthApi = createHook();
