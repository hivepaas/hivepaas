import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    RegistryAuth_CreateOne_Req,
    RegistryAuth_DeleteOne_Req,
    RegistryAuth_FindManyPaginated_Req,
    RegistryAuth_FindOneById_Req,
    RegistryAuth_TestConn_Req,
    RegistryAuth_UpdateMeta_Req,
    RegistryAuth_UpdateOne_Req,
} from "~/settings/api/services/registry-auth-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useRegistryAuthApi() {
        const { api } = use(SettingsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: RegistryAuth_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.registryAuth.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get registry auth settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: RegistryAuth_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.registryAuth.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get registry auth",
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
                createOne: async (data: RegistryAuth_CreateOne_Req["data"]) => {
                    const result = await api.settings.registryAuth.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create registry auth",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: RegistryAuth_UpdateOne_Req["data"]) => {
                    const result = await api.settings.registryAuth.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update registry auth",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateMeta: async (data: RegistryAuth_UpdateMeta_Req["data"]) => {
                    const result = await api.settings.registryAuth.updateMeta({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update registry auth meta",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: RegistryAuth_DeleteOne_Req["data"]) => {
                    const result = await api.settings.registryAuth.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete registry auth",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                testConn: async (data: RegistryAuth_TestConn_Req["data"]) => {
                    const result = await api.settings.registryAuth.testConn({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to test registry connection",
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

export const useRegistryAuthApi = createHook();
