import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { SystemSettingsApiContext } from "~/system-settings/api/api-context";
import type { HivePaaSUpdates_FindPlan_Req, HivePaaSUpdates_Update_Req } from "~/system-settings/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useHivePaaSUpdatesApi() {
        const { api } = use(SystemSettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /** quiet: fail without telling anyone, for a page that only mentions updates in passing. */
                findReleaseInfo: async (signal?: AbortSignal, quiet = false) => {
                    const result = await api.systemSettings.hivepaasUpdates.findReleaseInfo({ data: {} }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            if (!quiet) {
                                notifyError({ message: "Failed to check for HivePaaS updates", error });
                            }
                            throw error;
                        },
                    });
                },
                findPlan: async (data: HivePaaSUpdates_FindPlan_Req["data"], signal?: AbortSignal) => {
                    const result = await api.systemSettings.hivepaasUpdates.findPlan({ data }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to read what the update would do", error });
                            throw error;
                        },
                    });
                },
                // Answers false rather than failing: not answering is the expected state mid-update.
                ping: () => api.systemSettings.hivepaasUpdates.ping(),
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                update: async (data: HivePaaSUpdates_Update_Req["data"]) => {
                    const result = await api.systemSettings.hivepaasUpdates.update({ data });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to start the update", error });
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

export const useHivePaaSUpdatesApi = createHook();
