import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type { HivePaaSAppSecret_UpdateOne_Req } from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useHivePaaSAppSecretApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const mutations = useMemo(
            () => ({
                updateOne: async (data: HivePaaSAppSecret_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasAppSecret.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update HivePaaS app secret", error });
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

export const useHivePaaSAppSecretApi = createHook();
