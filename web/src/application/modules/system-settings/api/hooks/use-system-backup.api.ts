import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    SystemBackup_Execute_Req,
    SystemBackup_FindOne_Req,
    SystemBackup_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemBackupApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: SystemBackup_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.backup.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get system backup settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: SystemBackup_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.backup.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update system backup settings", error });
                            throw error;
                        },
                    });
                },
                execute: async (data: SystemBackup_Execute_Req["data"]) => {
                    const result = await api.systemSettings.backup.execute({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to execute system backup", error });
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

export const useSystemBackupApi = createHook();
