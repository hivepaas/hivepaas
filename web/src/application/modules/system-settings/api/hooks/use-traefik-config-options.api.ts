import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    TraefikConfigOptions_ConfirmChange_Req,
    TraefikConfigOptions_FindOne_Req,
    TraefikConfigOptions_RevertChange_Req,
    TraefikConfigOptions_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

import { isProbationErrorHandledByCaller } from "./settings-probation.errors";

function createHook() {
    return function useTraefikConfigOptionsApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (data: TraefikConfigOptions_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.traefikConfigOptions.findOne({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get Traefik config options", error });
                            throw error;
                        },
                    });
                },

                /**
                 * Same request as findOne, deliberately silent.
                 *
                 * It runs every few seconds while a command change is on trial, and
                 * its failures are the thing the dialog is displaying - they mean
                 * the new Traefik is not letting the caller back in. A toast every
                 * three seconds would pile noise on top of that.
                 */
                probe: async (signal?: AbortSignal) => {
                    const result = await api.systemSettings.traefikConfigOptions.findOne({ data: {} }, signal);

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
                updateOne: async (data: TraefikConfigOptions_UpdateOne_Req["data"]) => {
                    const result = await api.systemSettings.traefikConfigOptions.updateOne({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update Traefik config options", error });
                            throw error;
                        },
                    });
                },

                confirmChange: async (data: TraefikConfigOptions_ConfirmChange_Req["data"]) => {
                    const result = await api.systemSettings.traefikConfigOptions.confirmChange({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isProbationErrorHandledByCaller(error)) {
                                notifyError({ message: "Failed to confirm the Traefik config change", error });
                            }
                            throw error;
                        },
                    });
                },

                revertChange: async (data: TraefikConfigOptions_RevertChange_Req["data"]) => {
                    const result = await api.systemSettings.traefikConfigOptions.revertChange({ data });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!isProbationErrorHandledByCaller(error)) {
                                notifyError({ message: "Failed to revert the Traefik config change", error });
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

export const useTraefikConfigOptionsApi = createHook();
