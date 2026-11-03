import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    BackupRepo_Cleanup_Req,
    BackupRepo_CreateOne_Req,
    BackupRepo_DeleteOne_Req,
    BackupRepo_FindManyPaginated_Req,
    BackupRepo_FindOneById_Req,
    BackupRepo_Sync_Req,
    BackupRepo_UpdateOne_Req,
    BackupRepo_UpdatePassword_Req,
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
                createOne: async (data: BackupRepo_CreateOne_Req["data"]) => {
                    const result = await api.settings.backupRepo.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create backup repository",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: BackupRepo_UpdateOne_Req["data"]) => {
                    const result = await api.settings.backupRepo.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update backup repository",
                                error,
                            });

                            throw error;
                        },
                    });
                },
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
                updatePassword: async (data: BackupRepo_UpdatePassword_Req["data"]) => {
                    const result = await api.settings.backupRepo.updatePassword({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update backup repository password",
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
                cleanup: async (data: BackupRepo_Cleanup_Req["data"]) => {
                    const result = await api.settings.backupRepo.cleanup({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to cleanup backup repository",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                sync: async (data: BackupRepo_Sync_Req["data"]) => {
                    const result = await api.settings.backupRepo.sync({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to sync backup repository",
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
