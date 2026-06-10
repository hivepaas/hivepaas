import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { AppLogsWsHandlers, AppLogsWs_StreamLogs_Req } from "~/projects/api/services";

function createHook() {
    return function useAppLogsWsApi() {
        const { api } = use(ProjectsApiContext);

        const streams = useMemo(
            () => ({
                subscribe: async (
                    data: AppLogsWs_StreamLogs_Req["data"],
                    handlers: AppLogsWsHandlers,
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.apps.logs.stream.$.streamLogs({ data }, handlers, signal);

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

        return {
            streams,
        };
    };
}

export const useAppLogsWsApi = createHook();
