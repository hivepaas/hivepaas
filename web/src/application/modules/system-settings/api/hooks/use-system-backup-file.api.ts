import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    SystemBackupFile_DeleteOne_Req,
    SystemBackupFile_DownloadOne_Req,
    SystemBackupFile_FindManyPaginated_Req,
    SystemBackupFile_FindOneById_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemBackupFileApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: SystemBackupFile_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.systemSettings.backupFiles.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get backup files", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: SystemBackupFile_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.backupFiles.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get backup file", error });
                            throw error;
                        },
                    });
                },
                downloadOne: async (data: SystemBackupFile_DownloadOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.backupFiles.downloadOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to download backup file", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                deleteOne: async (data: SystemBackupFile_DeleteOne_Req["data"]) => {
                    const result = await api.systemSettings.backupFiles.deleteOne({ data });

                    return match(result, {
                        Ok: response => response,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete backup file",
                                error,
                            });

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

export const useSystemBackupFileApi = createHook();
