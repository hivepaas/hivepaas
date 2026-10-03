import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppLogs_GetDependencyMetrics_Req,
    AppLogs_GetFunctionMetrics_Req,
    AppLogs_GetHistory_Req,
    AppLogs_GetHttpMetrics_Req,
    AppLogs_GetInfo_Req,
    AppLogs_GetLogs_Req,
    AppLogs_GetResourceMetrics_Req,
    AppLogs_GetRouteMetrics_Req,
} from "~/projects/api/services";

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
                getResourceMetrics: async (data: AppLogs_GetResourceMetrics_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getResourceMetrics({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the app's CPU and memory",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                getHttpMetrics: async (data: AppLogs_GetHttpMetrics_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getHttpMetrics({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the app's HTTP metrics",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                getRouteMetrics: async (data: AppLogs_GetRouteMetrics_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getRouteMetrics({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the app's routes",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                getDependencyMetrics: async (data: AppLogs_GetDependencyMetrics_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getDependencyMetrics({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get what the app calls",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                getFunctionMetrics: async (data: AppLogs_GetFunctionMetrics_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getFunctionMetrics({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get the function's metrics",
                                error,
                            });

                            throw error;
                        },
                    });
                },
                getHistory: async (data: AppLogs_GetHistory_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.logs.$.getHistory({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get stored logs",
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
