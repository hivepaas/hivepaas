import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    SystemCleanup_Execute_Req,
    SystemCleanup_FindOne_Req,
    SystemCleanup_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemCleanupApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: SystemCleanup_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.cleanup.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get system cleanup settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: SystemCleanup_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.cleanup.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update system cleanup settings", error });
                            throw error;
                        },
                    });
                },
                execute: async (data: SystemCleanup_Execute_Req["data"]) => {
                    const result = await api.systemSettings.cleanup.execute({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to execute system cleanup", error });
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

export const useSystemCleanupApi = createHook();
