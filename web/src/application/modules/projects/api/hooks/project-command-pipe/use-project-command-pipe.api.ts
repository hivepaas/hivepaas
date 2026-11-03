import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectCommandPipe_CreateFromTemplate_Req,
    ProjectCommandPipe_CreateOne_Req,
    ProjectCommandPipe_DeleteOne_Req,
    ProjectCommandPipe_FindManyPaginated_Req,
    ProjectCommandPipe_FindOneById_Req,
    ProjectCommandPipe_UpdateOne_Req,
    ProjectCommandPipe_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectCommandPipeApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectCommandPipe_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.commandPipes.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project Command Pipes",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectCommandPipe_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.commandPipes.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project Command Pipe",
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
                createOne: async (data: ProjectCommandPipe_CreateOne_Req["data"]) => {
                    const result = await api.projects.commandPipes.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project Command Pipe",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                createFromTemplate: async (data: ProjectCommandPipe_CreateFromTemplate_Req["data"]) => {
                    const result = await api.projects.commandPipes.$.createFromTemplate({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project Command Pipe from template",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectCommandPipe_UpdateOne_Req["data"]) => {
                    const result = await api.projects.commandPipes.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project Command Pipe",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectCommandPipe_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.commandPipes.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project Command Pipe status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectCommandPipe_DeleteOne_Req["data"]) => {
                    const result = await api.projects.commandPipes.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project Command Pipe",
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

export const useProjectCommandPipeApi = createHook();
