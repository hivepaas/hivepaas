import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectImService_CreateOne_Req,
    ProjectImService_DeleteOne_Req,
    ProjectImService_FindManyPaginated_Req,
    ProjectImService_FindOneById_Req,
    ProjectImService_UpdateOne_Req,
    ProjectImService_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectImServiceApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectImService_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.imService.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project IM platforms",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectImService_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.imService.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project IM platform",
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
                createOne: async (data: ProjectImService_CreateOne_Req["data"]) => {
                    const result = await api.projects.imService.$.createOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to create project IM platform",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectImService_UpdateOne_Req["data"]) => {
                    const result = await api.projects.imService.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project IM platform",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectImService_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.imService.$.updateStatus({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project IM platform status",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectImService_DeleteOne_Req["data"]) => {
                    const result = await api.projects.imService.$.deleteOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to delete project IM platform",
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

export const useProjectImServiceApi = createHook();
