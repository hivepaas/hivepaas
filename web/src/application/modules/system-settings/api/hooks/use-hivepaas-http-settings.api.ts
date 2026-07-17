import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    HivePaaSHttpSettings_FindOne_Req,
    HivePaaSHttpSettings_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useHivePaaSHttpSettingsApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: HivePaaSHttpSettings_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasHttpSettings.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get HivePaaS HTTP settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: HivePaaSHttpSettings_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasHttpSettings.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update HivePaaS HTTP settings", error });
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

export const useHivePaaSHttpSettingsApi = createHook();
