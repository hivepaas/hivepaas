import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppSecrets_BuildDownloadUrl_Req,
    AppSecrets_CreateOne_Req,
    AppSecrets_DeleteOne_Req,
    AppSecrets_FindManyPaginated_Req,
    AppSecrets_FindOneById_Req,
    AppSecrets_GetDownloadToken_Req,
    AppSecrets_UpdateOne_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectAppSecretsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many app secrets paginated
                 */
                findManyPaginated: async (data: AppSecrets_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.secrets.$.findManyPaginated(
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
                 * Find one app secret by id
                 */
                findOneById: async (data: AppSecrets_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.secrets.$.findOneById(
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
                 * Get app secret download token
                 */
                getDownloadToken: async (data: AppSecrets_GetDownloadToken_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.secrets.$.getDownloadToken(
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

        const helpers = useMemo(
            () => ({
                /**
                 * Build app secret download URL
                 */
                buildDownloadUrl: (data: AppSecrets_BuildDownloadUrl_Req) => {
                    return api.projects.apps.secrets.$.buildDownloadUrl(data);
                },
            }),
            [api],
        );

        const mutations = useMemo(
            () => ({
                /**
                 * Create app secret
                 */
                createOne: async (data: AppSecrets_CreateOne_Req["data"]) => {
                    const result = await api.projects.apps.secrets.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create app secret",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Delete app secret
                 */
                deleteOne: async (data: AppSecrets_DeleteOne_Req["data"]) => {
                    const result = await api.projects.apps.secrets.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete app secret",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                /**
                 * Update app secret
                 */
                updateOne: async (data: AppSecrets_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.secrets.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update app secret",
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
            helpers,
            mutations,
        };
    };
}

export const useProjectAppSecretsApi = createHook();
