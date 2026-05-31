import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppScheduledJobs_CreateOne_Req,
    AppScheduledJobs_DeleteOne_Req,
    AppScheduledJobs_FindManyPaginated_Req,
    AppScheduledJobs_FindOneById_Req,
    AppScheduledJobs_RunNow_Req,
    AppScheduledJobs_UpdateOne_Req,
    AppScheduledJobs_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppScheduledJobsApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: AppScheduledJobs_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.apps.scheduledJobs.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
                findOneById: async (data: AppScheduledJobs_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.scheduledJobs.$.findOneById({ data }, signal);

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
                createOne: async (data: AppScheduledJobs_CreateOne_Req["data"]) => {
                    const result = await api.projects.apps.scheduledJobs.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create app scheduled job", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: AppScheduledJobs_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.scheduledJobs.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update app scheduled job", error });
                            throw error;
                        },
                    });
                },
                updateStatus: async (data: AppScheduledJobs_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.apps.scheduledJobs.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update app scheduled job status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: AppScheduledJobs_DeleteOne_Req["data"]) => {
                    const result = await api.projects.apps.scheduledJobs.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete app scheduled job", error });
                            throw error;
                        },
                    });
                },
                runNow: async (data: AppScheduledJobs_RunNow_Req["data"]) => {
                    const result = await api.projects.apps.scheduledJobs.$.runNow({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to run app scheduled job", error });
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

export const useAppScheduledJobsApi = createHook();
