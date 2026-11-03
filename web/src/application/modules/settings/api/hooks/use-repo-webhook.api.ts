import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    RepoWebhook_CreateOne_Req,
    RepoWebhook_DeleteOne_Req,
    RepoWebhook_FindManyPaginated_Req,
    RepoWebhook_FindOneById_Req,
    RepoWebhook_UpdateOne_Req,
    RepoWebhook_UpdateStatus_Req,
} from "~/settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useRepoWebhookApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: RepoWebhook_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.repoWebhook.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get repo webhook settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: RepoWebhook_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.repoWebhook.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get repo webhook setting",
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
                createOne: async (data: RepoWebhook_CreateOne_Req["data"]) => {
                    const result = await api.settings.repoWebhook.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create repo webhook setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: RepoWebhook_UpdateOne_Req["data"]) => {
                    const result = await api.settings.repoWebhook.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update repo webhook setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: RepoWebhook_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.repoWebhook.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update repo webhook status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: RepoWebhook_DeleteOne_Req["data"]) => {
                    const result = await api.settings.repoWebhook.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete repo webhook setting",
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

export const useRepoWebhookApi = createHook();
