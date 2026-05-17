import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    ProjectNotification_CreateOne_Req,
    ProjectNotification_DeleteOne_Req,
    ProjectNotification_FindManyPaginated_Req,
    ProjectNotification_FindOneById_Req,
    ProjectNotification_UpdateOne_Req,
    ProjectNotification_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectNotificationsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: ProjectNotification_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.notifications.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project notifications",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                findOneById: async (data: ProjectNotification_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.notifications.$.findOneById(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project notification",
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
                createOne: async (data: ProjectNotification_CreateOne_Req["data"]) => {
                    const result = await api.projects.notifications.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create project notification", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: ProjectNotification_UpdateOne_Req["data"]) => {
                    const result = await api.projects.notifications.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project notification", error });
                            throw error;
                        },
                    });
                },
                updateStatus: async (data: ProjectNotification_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.notifications.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update project notification status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: ProjectNotification_DeleteOne_Req["data"]) => {
                    const result = await api.projects.notifications.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete project notification", error });
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

export const useProjectNotificationsApi = createHook();
