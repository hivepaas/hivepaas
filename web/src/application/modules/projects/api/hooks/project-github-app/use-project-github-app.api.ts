import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectGithubApp_BeginManifestFlow_Req,
    ProjectGithubApp_BeginReprovision_Req,
    ProjectGithubApp_CreateOne_Req,
    ProjectGithubApp_DeleteOne_Req,
    ProjectGithubApp_FindManyPaginated_Req,
    ProjectGithubApp_FindOneById_Req,
    ProjectGithubApp_UpdateOne_Req,
    ProjectGithubApp_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectGithubAppApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectGithubApp_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.githubApp.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project github app settings",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectGithubApp_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.githubApp.$.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project github app setting",
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
                createOne: async (data: ProjectGithubApp_CreateOne_Req["data"]) => {
                    const result = await api.projects.githubApp.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project github app setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectGithubApp_UpdateOne_Req["data"]) => {
                    const result = await api.projects.githubApp.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project github app setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectGithubApp_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.githubApp.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project github app status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectGithubApp_DeleteOne_Req["data"]) => {
                    const result = await api.projects.githubApp.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project github app setting",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                beginManifestFlow: async (data: ProjectGithubApp_BeginManifestFlow_Req["data"]) => {
                    const result = await api.projects.githubApp.$.beginManifestFlow({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to begin project github app creation flow",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                beginReprovision: async (data: ProjectGithubApp_BeginReprovision_Req["data"]) => {
                    const result = await api.projects.githubApp.$.beginReprovision({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to begin project github app reprovision flow",
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

export const useProjectGithubAppApi = createHook();
