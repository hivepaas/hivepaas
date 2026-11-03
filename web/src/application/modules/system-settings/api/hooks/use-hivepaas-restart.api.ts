import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type { HivePaaSRestart_Execute_Req } from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useHivePaaSRestartApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                execute: async (data: HivePaaSRestart_Execute_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRestart.execute({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to restart HivePaaS services", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { mutations };
    };
}

export const useHivePaaSRestartApi = createHook();
