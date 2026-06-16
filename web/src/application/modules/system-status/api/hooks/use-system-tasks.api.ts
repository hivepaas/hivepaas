import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemStatusApiContext } from "~/system-status/api/api-context";
import type {
    SystemTasks_Cancel_Req,
    SystemTasks_FindManyPaginated_Req,
    SystemTasks_FindOneById_Req,
} from "~/system-status/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemTasksApi() {
        const { api } = use(SystemStatusApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: SystemTasks_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemStatus.tasks.findManyPaginated({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get system tasks", error });
                            throw error;
                        },
                    });
                },
                findOneById: async (data: SystemTasks_FindOneById_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemStatus.tasks.findOneById({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get system task", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                cancel: async (data: SystemTasks_Cancel_Req["data"]) => {
                    const result = await api.systemStatus.tasks.cancel({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to cancel system task", error });
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

export const useSystemTasksApi = createHook();
