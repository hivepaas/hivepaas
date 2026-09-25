import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppSettingMounts_CreateOne_Req,
    AppSettingMounts_DeleteOne_Req,
    AppSettingMounts_FindManyPaginated_Req,
    AppSettingMounts_FindOneById_Req,
    AppSettingMounts_FindSources_Req,
    AppSettingMounts_UpdateOne_Req,
    AppSettingMounts_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppSettingMountsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many app setting mounts paginated
                 */
                findManyPaginated: async (
                    data: AppSettingMounts_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.apps.settingMounts.$.findManyPaginated(
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
                 * Find one app setting mount by id
                 */
                findOneById: async (data: AppSettingMounts_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.settingMounts.$.findOneById(
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
                 * Find the setting types an entry may mount from
                 */
                findSources: async (data: AppSettingMounts_FindSources_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.settingMounts.$.findSources(
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
                 * Create app setting mount
                 */
                createOne: async (data: AppSettingMounts_CreateOne_Req["data"]) => {
                    const result = await api.projects.apps.settingMounts.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create setting mount",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update app setting mount
                 */
                updateOne: async (data: AppSettingMounts_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.settingMounts.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update setting mount",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Enable or disable app setting mount
                 */
                updateStatus: async (data: AppSettingMounts_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.apps.settingMounts.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update the status of setting mount",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete app setting mount
                 */
                deleteOne: async (data: AppSettingMounts_DeleteOne_Req["data"]) => {
                    const result = await api.projects.apps.settingMounts.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete setting mount",
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

export const useAppSettingMountsApi = createHook();
