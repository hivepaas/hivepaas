import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    Projects_CreateOne_Req,
    Projects_DeleteOne_Req,
    Projects_FindManyPaginated_Req,
    Projects_FindOneById_Req,
    Projects_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many projects paginated
                 */
                findManyPaginated: async (data: Projects_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get projects",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Find one project by id
                 */
                findOneById: async (data: Projects_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project",
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
                 * Create a project
                 */
                createOne: async (data: Projects_CreateOne_Req["data"]) => {
                    const result = await api.projects.$.createOne(
                        {
                            data,
                        },
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete a project
                 */
                deleteOne: async (data: Projects_DeleteOne_Req["data"]) => {
                    const result = await api.projects.$.deleteOne(
                        {
                            data,
                        },
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update a project
                 */
                updateOne: async (data: Projects_UpdateOne_Req["data"]) => {
                    const result = await api.projects.$.updateOne(
                        {
                            data,
                        },
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project",
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

export const useProjectsApi = createHook();
