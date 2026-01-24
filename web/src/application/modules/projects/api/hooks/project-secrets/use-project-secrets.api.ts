import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectSecrets_CreateOne_Req,
    ProjectSecrets_DeleteOne_Req,
    ProjectSecrets_FindManyPaginated_Req,
    ProjectSecrets_FindOneById_Req,
    ProjectSecrets_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectSecretsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many project secrets paginated
                 */
                findManyPaginated: async (data: ProjectSecrets_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.secrets.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project secrets",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Find one project secret by id
                 */
                findOneById: async (data: ProjectSecrets_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.secrets.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project secret",
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
                 * Create a project secret
                 */
                createOne: async (data: ProjectSecrets_CreateOne_Req["data"]) => {
                    const result = await api.projects.secrets.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project secret",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete a project secret
                 */
                deleteOne: async (data: ProjectSecrets_DeleteOne_Req["data"]) => {
                    const result = await api.projects.secrets.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project secret",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update a project secret
                 */
                updateOne: async (data: ProjectSecrets_UpdateOne_Req["data"]) => {
                    const result = await api.projects.secrets.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project secret",
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

export const useProjectSecretsApi = createHook();
