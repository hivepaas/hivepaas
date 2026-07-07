import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectCommandTemplate_CreateOne_Req,
    ProjectCommandTemplate_DeleteOne_Req,
    ProjectCommandTemplate_FindManyPaginated_Req,
    ProjectCommandTemplate_FindOneById_Req,
    ProjectCommandTemplate_UpdateOne_Req,
    ProjectCommandTemplate_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectCommandTemplateApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectCommandTemplate_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.commandTemplates.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project Command Templates",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectCommandTemplate_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.commandTemplates.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project Command Template",
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
                createOne: async (data: ProjectCommandTemplate_CreateOne_Req["data"]) => {
                    const result = await api.projects.commandTemplates.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project Command Template",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectCommandTemplate_UpdateOne_Req["data"]) => {
                    const result = await api.projects.commandTemplates.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project Command Template",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectCommandTemplate_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.commandTemplates.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project Command Template status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectCommandTemplate_DeleteOne_Req["data"]) => {
                    const result = await api.projects.commandTemplates.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project Command Template",
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

export const useProjectCommandTemplateApi = createHook();
