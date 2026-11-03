import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    SystemBackupRepoCleanup_Execute_Req,
    SystemBackupRepoCleanup_FindOne_Req,
    SystemBackupRepoCleanup_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemBackupRepoCleanupApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: SystemBackupRepoCleanup_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.backupRepoCleanup.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get backup repo cleanup settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: SystemBackupRepoCleanup_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.backupRepoCleanup.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update backup repo cleanup settings", error });
                            throw error;
                        },
                    });
                },
                execute: async (data: SystemBackupRepoCleanup_Execute_Req["data"]) => {
                    const result = await api.systemSettings.backupRepoCleanup.execute({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to execute backup repo cleanup", error });
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

export const useSystemBackupRepoCleanupApi = createHook();
