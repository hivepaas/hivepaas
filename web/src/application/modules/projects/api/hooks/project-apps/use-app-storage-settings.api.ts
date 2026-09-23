import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import { type AppStorageSettings_Preflight_Req, type AppStorageSettings_UpdateOne_Req } from "../../../api/services";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useAppStorageSettingsApi() {
        const context = use(ProjectsApiContext);
        const { api } = context;
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (request: { projectID: string; env: string; appID: string }, signal?: AbortSignal) => {
                    const result = await api.projects.apps.storageSettings.$.findOne({ data: request }, signal);
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            throw error;
                        },
                    });
                },
            }),
            [api],
        );

        const mutations = useMemo(
            () => ({
                updateOne: async (request: AppStorageSettings_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.storageSettings.$.updateOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update storage settings", error });
                            throw error;
                        },
                    });
                },
                // The check is advisory and reports nothing of its own: a failure
                // is handled where it is asked for, by saving anyway.
                preflight: async (request: AppStorageSettings_Preflight_Req["data"]) => {
                    const result = await api.projects.apps.storageSettings.$.preflight({ data: request });
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

        return { queries, mutations };
    };
}

export const useAppStorageSettingsApi = createHook();
