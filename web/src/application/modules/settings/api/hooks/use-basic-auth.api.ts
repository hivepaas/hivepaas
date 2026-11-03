import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    BasicAuth_CreateOne_Req,
    BasicAuth_DeleteOne_Req,
    BasicAuth_FindManyPaginated_Req,
    BasicAuth_FindOneById_Req,
    BasicAuth_UpdateOne_Req,
    BasicAuth_UpdateStatus_Req,
} from "~/settings/api/services/basic-auth-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useBasicAuthApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: BasicAuth_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.basicAuth.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get basic auth settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: BasicAuth_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.basicAuth.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get basic auth setting",
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
                createOne: async (data: BasicAuth_CreateOne_Req["data"]) => {
                    const result = await api.settings.basicAuth.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create basic auth setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: BasicAuth_UpdateOne_Req["data"]) => {
                    const result = await api.settings.basicAuth.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update basic auth setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: BasicAuth_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.basicAuth.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update basic auth status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: BasicAuth_DeleteOne_Req["data"]) => {
                    const result = await api.settings.basicAuth.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete basic auth setting",
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

export const useBasicAuthApi = createHook();
