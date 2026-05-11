import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppConfigFiles_CreateOne_Req,
    AppConfigFiles_DeleteOne_Req,
    AppConfigFiles_FindManyPaginated_Req,
    AppConfigFiles_FindOneById_Req,
    AppConfigFiles_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppConfigFilesApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many app config files paginated
                 */
                findManyPaginated: async (data: AppConfigFiles_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.configFiles.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                /**
                 * Find one app config file by id
                 */
                findOneById: async (data: AppConfigFiles_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.configFiles.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        const mutations = useMemo(
            () => ({
                /**
                 * Create app config file
                 */
                createOne: async (data: AppConfigFiles_CreateOne_Req["data"]) => {
                    const result = await api.projects.apps.configFiles.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create app config file",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete app config file
                 */
                deleteOne: async (data: AppConfigFiles_DeleteOne_Req["data"]) => {
                    const result = await api.projects.apps.configFiles.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete app config file",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update app config file
                 */
                updateOne: async (data: AppConfigFiles_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.configFiles.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update app config file",
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

export const useAppConfigFilesApi = createHook();
