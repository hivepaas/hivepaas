import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { AppLogs_GetInfo_Req, AppLogs_GetLogs_Req } from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useAppLogsApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                getInfo: async (data: AppLogs_GetInfo_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getInfo({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get app logs info",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                getLogs: async (data: AppLogs_GetLogs_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getLogs({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get app logs",
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
        };
    };
}

export const useAppLogsApi = createHook();
