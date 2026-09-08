import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    HivePaaSServiceSettings_ConfirmChange_Req,
    HivePaaSServiceSettings_FindOne_Req,
    HivePaaSServiceSettings_RevertChange_Req,
    HivePaaSServiceSettings_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

import { isProbationErrorHandledByCaller } from "./settings-probation.errors";

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

                /**
                 * Same request as findOne, deliberately silent.
                 *
                 * It runs every few seconds while a change is on trial, and its
                 * failures are what the dialog is displaying - during a proxy
                 * change they mean traefik is still restarting. A toast every
                 * three seconds would pile noise on top of that.
                 */
                probe: async (signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasServiceSettings.findOne({ data: {} }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
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

                confirmChange: async (data: HivePaaSServiceSettings_ConfirmChange_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasServiceSettings.confirmChange({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isProbationErrorHandledByCaller(error)) {
                                notifyError({ message: "Failed to confirm the service settings change", error });
                            }
                            throw error;
                        },
                    });
                },

                revertChange: async (data: HivePaaSServiceSettings_RevertChange_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasServiceSettings.revertChange({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isProbationErrorHandledByCaller(error)) {
                                notifyError({ message: "Failed to revert the service settings change", error });
                            }
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
