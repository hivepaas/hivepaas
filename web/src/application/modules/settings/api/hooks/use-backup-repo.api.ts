import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    BackupRepo_DeleteOne_Req,
    BackupRepo_FindManyPaginated_Req,
    BackupRepo_FindOneById_Req,
    BackupRepo_UpdateStatus_Req,
} from "~/settings/api/services/backup-repo-services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useBackupRepoApi() {
        const { api } = use(SettingsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: BackupRepo_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.backupRepo.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get backup repositories",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: BackupRepo_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.backupRepo.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get backup repository",
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
                updateStatus: async (data: BackupRepo_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.backupRepo.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update backup repository status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: BackupRepo_DeleteOne_Req["data"]) => {
                    const result = await api.settings.backupRepo.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete backup repository",
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

export const useBackupRepoApi = createHook();
