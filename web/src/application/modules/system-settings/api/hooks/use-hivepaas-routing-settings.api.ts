import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    HivePaaSRoutingSettings_ConfirmChange_Req,
    HivePaaSRoutingSettings_FindOne_Req,
    HivePaaSRoutingSettings_RevertChange_Req,
    HivePaaSRoutingSettings_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

import { HttpException } from "@infrastructure/exceptions/http";

/**
 * Errors the confirm dialog explains better than a toast can.
 *
 * "Too early" is not a failure at all - it is the proxy still catching up, and
 * the dialog answers it by waiting and retrying. A toast would turn a normal
 * step of the flow into something that looks broken.
 */
const CONFIRM_ERRORS_HANDLED_BY_CALLER = new Set([
    "ERR_SETTINGS_CONFIRM_TOO_EARLY",
    "ERR_SETTINGS_CHANGE_SUPERSEDED",
    "ERR_SETTINGS_NO_PENDING_CHANGE",
]);

function isHandledByCaller(error: Error): boolean {
    return error instanceof HttpException && CONFIRM_ERRORS_HANDLED_BY_CALLER.has(error.code);
}

function createHook() {
    return function useHivePaaSRoutingSettingsApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: HivePaaSRoutingSettings_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasRoutingSettings.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get HivePaaS Routing settings", error });
                            throw error;
                        },
                    });
                },

                /**
                 * Same request as findOne, deliberately silent.
                 *
                 * It runs every few seconds while a routing change is on trial, and
                 * its failures are the thing the dialog is displaying - they mean
                 * the new configuration is not letting the caller back in. A toast
                 * every three seconds would pile noise on top of that.
                 */
                probe: async (signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasRoutingSettings.findOne({ data: {} }, signal);

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
                updateOne: async (data: HivePaaSRoutingSettings_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRoutingSettings.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update HivePaaS Routing settings", error });
                            throw error;
                        },
                    });
                },

                confirmChange: async (data: HivePaaSRoutingSettings_ConfirmChange_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRoutingSettings.confirmChange({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isHandledByCaller(error)) {
                                notifyError({ message: "Failed to confirm the routing change", error });
                            }
                            throw error;
                        },
                    });
                },

                revertChange: async (data: HivePaaSRoutingSettings_RevertChange_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasRoutingSettings.revertChange({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isHandledByCaller(error)) {
                                notifyError({ message: "Failed to revert the routing change", error });
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

export const useHivePaaSRoutingSettingsApi = createHook();
export { useHivePaaSRoutingSettingsApi as useHivePaaSHttpSettingsApi };
