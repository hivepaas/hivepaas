import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    HivePaaSServiceSettings_FindOne_Req,
    HivePaaSServiceSettings_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useHivePaaSServiceSettingsApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: HivePaaSServiceSettings_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasServiceSettings.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get HivePaaS service settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: HivePaaSServiceSettings_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasServiceSettings.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update HivePaaS service settings", error });
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

export const useHivePaaSServiceSettingsApi = createHook();
