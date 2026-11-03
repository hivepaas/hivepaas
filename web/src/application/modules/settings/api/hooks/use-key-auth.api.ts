import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    KeyAuth_CreateOne_Req,
    KeyAuth_DeleteOne_Req,
    KeyAuth_FindManyPaginated_Req,
    KeyAuth_FindOneById_Req,
    KeyAuth_UpdateOne_Req,
    KeyAuth_UpdateStatus_Req,
} from "~/settings/api/services/key-auth-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useKeyAuthApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: KeyAuth_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.keyAuth.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get key auth settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: KeyAuth_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.keyAuth.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get key auth setting",
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
                createOne: async (data: KeyAuth_CreateOne_Req["data"]) => {
                    const result = await api.settings.keyAuth.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create key auth setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: KeyAuth_UpdateOne_Req["data"]) => {
                    const result = await api.settings.keyAuth.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update key auth setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: KeyAuth_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.keyAuth.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update key auth status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: KeyAuth_DeleteOne_Req["data"]) => {
                    const result = await api.settings.keyAuth.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete key auth setting",
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

export const useKeyAuthApi = createHook();
