import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    EnvScheduledJobs_CreateOne_Req,
    EnvScheduledJobs_DeleteOne_Req,
    EnvScheduledJobs_FindManyPaginated_Req,
    EnvScheduledJobs_FindOneById_Req,
    EnvScheduledJobs_RunNow_Req,
    EnvScheduledJobs_UpdateOne_Req,
    EnvScheduledJobs_UpdateStatus_Req,
} from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useEnvScheduledJobsApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (
                    data: EnvScheduledJobs_FindManyPaginated_Req["data"],
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.envScheduledJobs.$.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to load scheduled jobs", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: EnvScheduledJobs_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.envScheduledJobs.$.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                createOne: async (data: EnvScheduledJobs_CreateOne_Req["data"]) => {
                    const result = await api.projects.envScheduledJobs.$.createOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to create job sequence", error });
                            throw error;
                        },
                    });
                },
                updateOne: async (data: EnvScheduledJobs_UpdateOne_Req["data"]) => {
                    const result = await api.projects.envScheduledJobs.$.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update job sequence", error });
                            throw error;
                        },
                    });
                },
                updateStatus: async (data: EnvScheduledJobs_UpdateStatus_Req["data"]) => {
                    const result = await api.projects.envScheduledJobs.$.updateStatus({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update job sequence status", error });
                            throw error;
                        },
                    });
                },
                deleteOne: async (data: EnvScheduledJobs_DeleteOne_Req["data"]) => {
                    const result = await api.projects.envScheduledJobs.$.deleteOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to delete job sequence", error });
                            throw error;
                        },
                    });
                },
                runNow: async (data: EnvScheduledJobs_RunNow_Req["data"]) => {
                    const result = await api.projects.envScheduledJobs.$.runNow({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to run job sequence", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useEnvScheduledJobsApi = createHook();
