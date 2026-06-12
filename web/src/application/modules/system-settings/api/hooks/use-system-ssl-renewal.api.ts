import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    SystemSslRenewal_Execute_Req,
    SystemSslRenewal_FindOne_Req,
    SystemSslRenewal_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemSslRenewalApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: SystemSslRenewal_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.sslRenewal.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get SSL renewal settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: SystemSslRenewal_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.sslRenewal.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update SSL renewal settings", error });
                            throw error;
                        },
                    });
                },
                execute: async (data: SystemSslRenewal_Execute_Req["data"]) => {
                    const result = await api.systemSettings.sslRenewal.execute({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to execute SSL renewal", error });
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

export const useSystemSslRenewalApi = createHook();
