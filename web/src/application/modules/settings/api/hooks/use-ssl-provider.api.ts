import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    SslProvider_CreateOne_Req,
    SslProvider_DeleteOne_Req,
    SslProvider_FindManyPaginated_Req,
    SslProvider_FindOneById_Req,
    SslProvider_UpdateOne_Req,
    SslProvider_UpdateStatus_Req,
} from "~/settings/api/services/ssl-provider-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSslProviderApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: SslProvider_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.sslProvider.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get SSL provider settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: SslProvider_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.sslProvider.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get SSL provider setting",
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
                createOne: async (data: SslProvider_CreateOne_Req["data"]) => {
                    const result = await api.settings.sslProvider.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create SSL provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: SslProvider_UpdateOne_Req["data"]) => {
                    const result = await api.settings.sslProvider.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update SSL provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: SslProvider_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.sslProvider.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update SSL provider status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: SslProvider_DeleteOne_Req["data"]) => {
                    const result = await api.settings.sslProvider.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete SSL provider setting",
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

export const useSslProviderApi = createHook();
