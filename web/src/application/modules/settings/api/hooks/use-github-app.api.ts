import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SettingsApiContext } from "~/settings/api/api-context/settings.api.context";
import type {
    GithubApp_BeginManifestFlow_Req,
    GithubApp_BeginReprovision_Req,
    GithubApp_CreateOne_Req,
    GithubApp_DeleteOne_Req,
    GithubApp_FindManyPaginated_Req,
    GithubApp_FindOneById_Req,
    GithubApp_ListInstallations_Req,
    GithubApp_TestConnection_Req,
    GithubApp_UpdateOne_Req,
    GithubApp_UpdateStatus_Req,
} from "~/settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useGithubAppApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: GithubApp_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.githubApp.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get github app settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: GithubApp_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.settings.githubApp.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get github app setting",
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
                createOne: async (data: GithubApp_CreateOne_Req["data"]) => {
                    const result = await api.settings.githubApp.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create github app setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: GithubApp_UpdateOne_Req["data"]) => {
                    const result = await api.settings.githubApp.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update github app setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: GithubApp_UpdateStatus_Req["data"]) => {
                    const result = await api.settings.githubApp.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update github app status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: GithubApp_DeleteOne_Req["data"]) => {
                    const result = await api.settings.githubApp.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete github app setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                testConnection: async (data: GithubApp_TestConnection_Req["data"]) => {
                    const result = await api.settings.githubApp.testConnection({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to test github app connection",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                beginManifestFlow: async (data: GithubApp_BeginManifestFlow_Req["data"]) => {
                    const result = await api.settings.githubApp.beginManifestFlow({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to begin github app creation flow",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                beginReprovision: async (data: GithubApp_BeginReprovision_Req["data"]) => {
                    const result = await api.settings.githubApp.beginReprovision({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to begin github app reprovision flow",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                listInstallations: async (data: GithubApp_ListInstallations_Req["data"]) => {
                    const result = await api.settings.githubApp.listInstallations({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get github app installations",
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

export const useGithubAppApi = createHook();
