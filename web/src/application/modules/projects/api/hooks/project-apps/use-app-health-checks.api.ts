import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppHealthChecks_CreateOne_Req,
    AppHealthChecks_DeleteOne_Req,
    AppHealthChecks_FindManyPaginated_Req,
    AppHealthChecks_FindOneById_Req,
    AppHealthChecks_UpdateOne_Req,
    AppHealthChecks_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppHealthChecksApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: AppHealthChecks_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.apps.healthChecks.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                findOneById: async (data: AppHealthChecks_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.healthChecks.$.findOneById({ data }, signal);

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

        const mutations = useMemo(
            () => ({
                createOne: async (data: AppHealthChecks_CreateOne_Req["data"]) => {
                    const result = await api.projects.apps.healthChecks.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create app health check", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: AppHealthChecks_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.healthChecks.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update app health check", error });
                            throw error;
                        },
                    });
                },
                updateStatus: async (data: AppHealthChecks_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.apps.healthChecks.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update app health check status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: AppHealthChecks_DeleteOne_Req["data"]) => {
                    const result = await api.projects.apps.healthChecks.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete app health check", error });
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

export const useAppHealthChecksApi = createHook();
