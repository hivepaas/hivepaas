import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectBackupRepo_Cleanup_Req,
    ProjectBackupRepo_CreateOne_Req,
    ProjectBackupRepo_DeleteOne_Req,
    ProjectBackupRepo_FindManyPaginated_Req,
    ProjectBackupRepo_FindOneById_Req,
    ProjectBackupRepo_Sync_Req,
    ProjectBackupRepo_UpdateOne_Req,
    ProjectBackupRepo_UpdatePassword_Req,
    ProjectBackupRepo_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectBackupRepoApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectBackupRepo_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.backupRepo.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project backup repositories",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectBackupRepo_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.backupRepo.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project backup repository",
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
                createOne: async (data: ProjectBackupRepo_CreateOne_Req["data"]) => {
                    const result = await api.projects.backupRepo.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project backup repository",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectBackupRepo_UpdateOne_Req["data"]) => {
                    const result = await api.projects.backupRepo.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project backup repository",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectBackupRepo_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.backupRepo.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project backup repository status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updatePassword: async (data: ProjectBackupRepo_UpdatePassword_Req["data"]) => {
                    const result = await api.projects.backupRepo.$.updatePassword({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project backup repository password",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectBackupRepo_DeleteOne_Req["data"]) => {
                    const result = await api.projects.backupRepo.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project backup repository",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                cleanup: async (data: ProjectBackupRepo_Cleanup_Req["data"]) => {
                    const result = await api.projects.backupRepo.$.cleanup({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to cleanup project backup repository",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                sync: async (data: ProjectBackupRepo_Sync_Req["data"]) => {
                    const result = await api.projects.backupRepo.$.sync({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to sync project backup repository",
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

export const useProjectBackupRepoApi = createHook();
