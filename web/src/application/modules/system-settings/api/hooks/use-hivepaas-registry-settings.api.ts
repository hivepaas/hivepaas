import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    HivePaaSRegistrySettings_CheckPush_Req,
    HivePaaSRegistrySettings_FindOne_Req,
    HivePaaSRegistrySettings_ProbeDomain_Req,
    HivePaaSRegistrySettings_RotateCredential_Req,
    HivePaaSRegistrySettings_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useHivePaaSRegistrySettingsApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: HivePaaSRegistrySettings_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasRegistrySettings.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get registry settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (data: HivePaaSRegistrySettings_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRegistrySettings.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update registry settings", error });
                            throw error;
                        },
                    });
                },
                probeDomain: async (data: HivePaaSRegistrySettings_ProbeDomain_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRegistrySettings.probeDomain({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to check the registry domain", error });
                            throw error;
                        },
                    });
                },
                checkPush: async (data: HivePaaSRegistrySettings_CheckPush_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRegistrySettings.checkPush({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to check a large push", error });
                            throw error;
                        },
                    });
                },
                rotateCredential: async (data: HivePaaSRegistrySettings_RotateCredential_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRegistrySettings.rotateCredential({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to rotate the registry password", error });
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

export const useHivePaaSRegistrySettingsApi = createHook();
