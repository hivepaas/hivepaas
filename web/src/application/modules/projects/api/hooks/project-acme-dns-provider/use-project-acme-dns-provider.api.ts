import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectAcmeDnsProvider_CreateOne_Req,
    ProjectAcmeDnsProvider_DeleteOne_Req,
    ProjectAcmeDnsProvider_FindManyPaginated_Req,
    ProjectAcmeDnsProvider_FindOneById_Req,
    ProjectAcmeDnsProvider_UpdateOne_Req,
    ProjectAcmeDnsProvider_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectAcmeDnsProviderApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectAcmeDnsProvider_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.acmeDnsProvider.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project ACME DNS provider settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectAcmeDnsProvider_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.acmeDnsProvider.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project ACME DNS provider setting",
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
                createOne: async (data: ProjectAcmeDnsProvider_CreateOne_Req["data"]) => {
                    const result = await api.projects.acmeDnsProvider.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project ACME DNS provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectAcmeDnsProvider_UpdateOne_Req["data"]) => {
                    const result = await api.projects.acmeDnsProvider.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project ACME DNS provider setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectAcmeDnsProvider_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.acmeDnsProvider.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project ACME DNS provider status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectAcmeDnsProvider_DeleteOne_Req["data"]) => {
                    const result = await api.projects.acmeDnsProvider.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project ACME DNS provider setting",
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

export const useProjectAcmeDnsProviderApi = createHook();
