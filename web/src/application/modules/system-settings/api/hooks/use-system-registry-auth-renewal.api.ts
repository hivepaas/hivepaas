import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    SystemRegistryAuthRenewal_Execute_Req,
    SystemRegistryAuthRenewal_FindOne_Req,
    SystemRegistryAuthRenewal_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useSystemRegistryAuthRenewalApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: SystemRegistryAuthRenewal_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.registryAuthRenewal.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get registry auth renewal settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: SystemRegistryAuthRenewal_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.registryAuthRenewal.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update registry auth renewal settings", error });
                            throw error;
                        },
                    });
                },
                execute: async (data: SystemRegistryAuthRenewal_Execute_Req["data"]) => {
                    const result = await api.systemSettings.registryAuthRenewal.execute({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to run the registry auth renewal", error });
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

export const useSystemRegistryAuthRenewalApi = createHook();
