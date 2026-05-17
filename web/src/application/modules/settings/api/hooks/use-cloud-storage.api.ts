import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    CloudStorage_CreateOne_Req,
    CloudStorage_DeleteOne_Req,
    CloudStorage_FindManyPaginated_Req,
    CloudStorage_FindOneById_Req,
    CloudStorage_TestConn_Req,
    CloudStorage_UpdateMeta_Req,
    CloudStorage_UpdateOne_Req,
} from "~/settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useCloudStorageApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: CloudStorage_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.cloudStorage.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get cloud storages", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: CloudStorage_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.cloudStorage.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get cloud storage", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: CloudStorage_CreateOne_Req["data"]) => {
                    const result = await api.settings.cloudStorage.createOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create cloud storage", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: CloudStorage_UpdateOne_Req["data"]) => {
                    const result = await api.settings.cloudStorage.updateOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update cloud storage", error });
                            throw error;
                        },
                    });
                },
                updateMeta: async (data: CloudStorage_UpdateMeta_Req["data"]) => {
                    const result = await api.settings.cloudStorage.updateMeta({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update cloud storage status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: CloudStorage_DeleteOne_Req["data"]) => {
                    const result = await api.settings.cloudStorage.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete cloud storage", error });
                            throw error;
                        },
                    });
                },
                testConn: async (data: CloudStorage_TestConn_Req["data"]) => {
                    const result = await api.settings.cloudStorage.testConn({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to test cloud storage", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useCloudStorageApi = createHook();
