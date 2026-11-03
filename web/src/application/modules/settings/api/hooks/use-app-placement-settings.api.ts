import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import { SettingsApiContext } from "../api-context/settings.api.context";
import type { AppPlacementSettings_UpdateOne_Req } from "../services";

function createHook() {
    return function useAppPlacementSettingsApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (signal?: AbortSignal) => {
                    const result = await api.settings.appPlacementSettings.findOne({ data: {} }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get app placement settings", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (request: AppPlacementSettings_UpdateOne_Req["data"]) => {
                    const result = await api.settings.appPlacementSettings.updateOne({ data: request });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update app placement settings", error });
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

export const useAppPlacementSettingsApi = createHook();
