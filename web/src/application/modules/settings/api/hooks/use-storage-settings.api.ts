import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import { SettingsApiContext } from "../api-context/settings.api.context";
import type {
    StorageSettings_DeleteOne_Req,
    StorageSettings_UpdateOne_Req,
    StorageSettings_UpdateStatus_Req,
} from "../services";

function createHook() {
    return function useStorageSettingsApi() {
        const context = use(SettingsApiContext);
        const { api } = context;
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (signal?: AbortSignal) => {
                    const result = await api.settings.storageSettings.findOne({ data: {} }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (request: StorageSettings_UpdateOne_Req["data"]) => {
                    const result = await api.settings.storageSettings.updateOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update storage settings", error });
                            throw error;
                        },
                    });
                },
                updateStatus: async (request: StorageSettings_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.storageSettings.updateStatus({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update storage settings status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (request: StorageSettings_DeleteOne_Req["data"]) => {
                    const result = await api.settings.storageSettings.deleteOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete storage settings", error });
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

export const useStorageSettingsApi = createHook();
