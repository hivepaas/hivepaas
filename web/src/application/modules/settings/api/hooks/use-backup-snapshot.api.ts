import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    BackupSnapshot_DeleteOne_Req,
    BackupSnapshot_DownloadFile_Req,
    BackupSnapshot_FindEntries_Req,
    BackupSnapshot_FindManyPaginated_Req,
    BackupSnapshot_FindOneById_Req,
    BackupSnapshot_Restore_Req,
} from "~/settings/api/services/backup-snapshot-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useBackupSnapshotApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: BackupSnapshot_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.backupSnapshot.findManyPaginated({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get backup snapshots", error });
                            throw error;
                        },
                    });
                },
                findEntries: async (data: BackupSnapshot_FindEntries_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.backupSnapshot.findEntries({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to list what the snapshot holds", error });
                            throw error;
                        },
                    });
                },
                downloadFile: async (data: BackupSnapshot_DownloadFile_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.backupSnapshot.downloadFile({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to download the file", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: BackupSnapshot_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.backupSnapshot.findOneById({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get backup snapshot", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                restore: async (data: BackupSnapshot_Restore_Req["data"]) => {
                    const result = await api.settings.backupSnapshot.restore({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to restore backup snapshot", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: BackupSnapshot_DeleteOne_Req["data"]) => {
                    const result = await api.settings.backupSnapshot.deleteOne({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete backup snapshot", error });
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

export const useBackupSnapshotApi = createHook();
