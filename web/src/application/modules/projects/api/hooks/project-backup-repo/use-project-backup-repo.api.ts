import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectBackupRepo_DeleteOne_Req,
    ProjectBackupRepo_FindManyPaginated_Req,
    ProjectBackupRepo_FindOneById_Req,
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
