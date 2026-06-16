import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemStatusApiContext } from "~/system-status/api/api-context";
import type { SystemTaskLogsWsHandlers, SystemTaskLogsWs_StreamLogs_Req } from "~/system-status/api/services";

function createHook() {
    return function useSystemTaskLogsWsApi() {
        const { api } = use(SystemStatusApiContext);

        const streams = useMemo(
            () => ({
                subscribe: async (
                    data: SystemTaskLogsWs_StreamLogs_Req["data"],
                    handlers: SystemTaskLogsWsHandlers,
                    signal?: AbortSignal,
                ) => {
                    const result = await Promise.resolve(
                        api.systemStatus.taskLogs.$.streamLogs({ data }, handlers, signal),
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

export const useSystemTaskLogsWsApi = createHook();
