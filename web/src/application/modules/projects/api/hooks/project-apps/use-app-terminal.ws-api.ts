import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { AppTerminalWsHandlers, AppTerminalWs_Open_Req } from "~/projects/api/services";

function createHook() {
    return function useAppTerminalWsApi() {
        const { api } = use(ProjectsApiContext);

        const streams = useMemo(
            () => ({
                open: async (
                    data: AppTerminalWs_Open_Req["data"],
                    handlers: AppTerminalWsHandlers,
                    signal?: AbortSignal,
                ) => {
                    const result = await Promise.resolve(
                        api.projects.apps.terminal.stream.$.open({ data }, handlers, signal),
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

export const useAppTerminalWsApi = createHook();
