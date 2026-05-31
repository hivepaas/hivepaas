import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectRepoWebhook_CreateOne_Req,
    ProjectRepoWebhook_DeleteOne_Req,
    ProjectRepoWebhook_FindManyPaginated_Req,
    ProjectRepoWebhook_FindOneById_Req,
    ProjectRepoWebhook_UpdateOne_Req,
    ProjectRepoWebhook_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectRepoWebhookApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectRepoWebhook_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.repoWebhook.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project repo webhook settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectRepoWebhook_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.repoWebhook.$.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project repo webhook setting",
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
                createOne: async (data: ProjectRepoWebhook_CreateOne_Req["data"]) => {
                    const result = await api.projects.repoWebhook.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project repo webhook setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectRepoWebhook_UpdateOne_Req["data"]) => {
                    const result = await api.projects.repoWebhook.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project repo webhook setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectRepoWebhook_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.repoWebhook.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project repo webhook status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectRepoWebhook_DeleteOne_Req["data"]) => {
                    const result = await api.projects.repoWebhook.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project repo webhook setting",
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

export const useProjectRepoWebhookApi = createHook();
