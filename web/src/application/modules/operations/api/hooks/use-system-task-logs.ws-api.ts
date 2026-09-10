import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { OperationsApiContext } from "~/operations/api/api-context";
import type { SystemTaskLogsWsHandlers, SystemTaskLogsWs_StreamLogs_Req } from "~/operations/api/services";

function createHook() {
    return function useSystemTaskLogsWsApi() {
        const { api } = use(OperationsApiContext);

        const streams = useMemo(
            () => ({
                subscribe: async (
                    data: SystemTaskLogsWs_StreamLogs_Req["data"],
                    handlers: SystemTaskLogsWsHandlers,
                    signal?: AbortSignal,
                ) => {
                    const result = await Promise.resolve(
                        api.operations.taskLogs.$.streamLogs({ data }, handlers, signal),
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
