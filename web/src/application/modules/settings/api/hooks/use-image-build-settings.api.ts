import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import { SettingsApiContext } from "../api-context/settings.api.context";
import type { ImageBuildSettings_ClearRepoCache_Req, ImageBuildSettings_UpdateOne_Req } from "../services";

function createHook() {
    return function useImageBuildSettingsApi() {
        const { api } = use(SettingsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (signal?: AbortSignal) => {
                    const result = await api.settings.imageBuildSettings.findOne({ data: {} }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get image build settings", error });
                            throw error;
                        },
                    });
                },
                findRepoCache: async (signal?: AbortSignal) => {
                    const result = await api.settings.imageBuildSettings.findRepoCache({ data: {} }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to get repo cache info", error });
                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (request: ImageBuildSettings_UpdateOne_Req["data"]) => {
                    const result = await api.settings.imageBuildSettings.updateOne({ data: request });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update image build settings", error });
                            throw error;
                        },
                    });
                },
                clearRepoCache: async (request: ImageBuildSettings_ClearRepoCache_Req["data"] = {}) => {
                    const result = await api.settings.imageBuildSettings.clearRepoCache({ data: request });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to clear repo cache", error });
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

export const useImageBuildSettingsApi = createHook();
