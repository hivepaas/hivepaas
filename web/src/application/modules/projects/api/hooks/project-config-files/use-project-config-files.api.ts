import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectConfigFiles_CreateOne_Req,
    ProjectConfigFiles_DeleteOne_Req,
    ProjectConfigFiles_FindManyPaginated_Req,
    ProjectConfigFiles_FindOneById_Req,
    ProjectConfigFiles_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectConfigFilesApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many project config files paginated
                 */
                findManyPaginated: async (
                    data: ProjectConfigFiles_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.configFiles.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project config files",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Find one project config file by id
                 */
                findOneById: async (data: ProjectConfigFiles_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.configFiles.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project config file",
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
                /**
                 * Create a project config file
                 */
                createOne: async (data: ProjectConfigFiles_CreateOne_Req["data"]) => {
                    const result = await api.projects.configFiles.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project config file",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete a project config file
                 */
                deleteOne: async (data: ProjectConfigFiles_DeleteOne_Req["data"]) => {
                    const result = await api.projects.configFiles.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project config file",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update a project config file
                 */
                updateOne: async (data: ProjectConfigFiles_UpdateOne_Req["data"]) => {
                    const result = await api.projects.configFiles.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project config file",
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

export const useProjectConfigFilesApi = createHook();
