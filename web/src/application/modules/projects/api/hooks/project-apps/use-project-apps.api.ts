import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectApps_Copy_Req,
    ProjectApps_CreateOne_Req,
    ProjectApps_DeleteOne_Req,
    ProjectApps_Deploy_Req,
    ProjectApps_FindManyPaginated_Req,
    ProjectApps_FindOneById_Req,
    ProjectApps_PrepareCopy_Req,
    ProjectApps_Restart_Req,
    ProjectApps_UpdateOne_Req,
    ProjectApps_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectAppsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many project apps paginated
                 */
                findManyPaginated: async (data: ProjectApps_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project apps",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Find one project app by id
                 */
                findOneById: async (data: ProjectApps_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project app",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                prepareCopy: async (data: ProjectApps_PrepareCopy_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.$.prepareCopy(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to prepare project app copy",
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
                 * Create a project app
                 */
                createOne: async (data: ProjectApps_CreateOne_Req["data"]) => {
                    const result = await api.projects.apps.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project app",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                copy: async (data: ProjectApps_Copy_Req["data"]) => {
                    const result = await api.projects.apps.$.copy({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to copy project app",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete a project app
                 */
                deleteOne: async (data: ProjectApps_DeleteOne_Req["data"]) => {
                    const result = await api.projects.apps.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project app",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update a project app
                 */
                updateOne: async (data: ProjectApps_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project app",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update a project app status
                 */
                updateStatus: async (data: ProjectApps_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.apps.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project app status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Deploy a project app using existing deployment settings
                 */
                deploy: async (data: ProjectApps_Deploy_Req["data"]) => {
                    const result = await api.projects.apps.$.deploy({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to re-deploy project app",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Restart a project app
                 */
                restart: async (data: ProjectApps_Restart_Req["data"]) => {
                    const result = await api.projects.apps.$.restart({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to restart project app",
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

export const useProjectAppsApi = createHook();
