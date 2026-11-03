import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type {
    AppScheduledJobTaskLogsWsHandlers,
    AppScheduledJobTaskLogsWs_StreamLogs_Req,
} from "~/projects/api/services";

function createHook() {
    return function useAppScheduledJobTaskLogsWsApi() {
        const { api } = use(ProjectsApiContext);

        const streams = useMemo(
            () => ({
                subscribe: async (
                    data: AppScheduledJobTaskLogsWs_StreamLogs_Req["data"],
                    handlers: AppScheduledJobTaskLogsWsHandlers,
                    signal?: AbortSignal,
                ) => {
                    const result = await Promise.resolve(
                        api.projects.apps.scheduledJobs.taskLogs.$.streamLogs({ data }, handlers, signal),
                    );

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

export const useAppScheduledJobTaskLogsWsApi = createHook();
