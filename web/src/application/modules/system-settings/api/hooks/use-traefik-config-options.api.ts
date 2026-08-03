import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type {
    TraefikConfigOptions_FindOne_Req,
    TraefikConfigOptions_UpdateOne_Req,
} from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

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
            }),
            [api, notifyError],
        );

        return { queries, mutations };
    };
}

export const useTraefikConfigOptionsApi = createHook();
