import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { AppDeploymentLogsWsHandlers, AppDeploymentLogsWs_StreamLogs_Req } from "~/projects/api/services";

function createHook() {
    return function useAppDeploymentLogsWsApi() {
        const { api } = use(ProjectsApiContext);

        const streams = useMemo(
            () => ({
                subscribe: async (
                    data: AppDeploymentLogsWs_StreamLogs_Req["data"],
                    handlers: AppDeploymentLogsWsHandlers,
                    signal?: AbortSignal,
                ) => {
                    const result = await api.projects.apps.deployments.logs.$.streamLogs({ data }, handlers, signal);

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

export const useAppDeploymentLogsWsApi = createHook();
