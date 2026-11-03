import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectEmail_CreateOne_Req,
    ProjectEmail_DeleteOne_Req,
    ProjectEmail_FindManyPaginated_Req,
    ProjectEmail_FindOneById_Req,
    ProjectEmail_UpdateOne_Req,
    ProjectEmail_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectEmailApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: ProjectEmail_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.email.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project Email accounts",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectEmail_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.email.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project Email account",
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
                createOne: async (data: ProjectEmail_CreateOne_Req["data"]) => {
                    const result = await api.projects.email.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project Email account",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectEmail_UpdateOne_Req["data"]) => {
                    const result = await api.projects.email.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project Email account",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectEmail_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.email.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project Email account status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectEmail_DeleteOne_Req["data"]) => {
                    const result = await api.projects.email.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project Email account",
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

export const useProjectEmailApi = createHook();
