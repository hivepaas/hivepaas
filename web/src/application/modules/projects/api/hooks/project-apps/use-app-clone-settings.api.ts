import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import type { AppCloneSettings_Execute_Req, AppCloneSettings_UpdateOne_Req } from "../../../api/services";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useAppCloneSettingsApi() {
        const context = use(ProjectsApiContext);
        const { api } = context;
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (request: { projectID: string; env: string; appID: string }, signal?: AbortSignal) => {
                    const result = await api.projects.apps.cloneSettings.$.findOne({ data: request }, signal);
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
                updateOne: async (request: AppCloneSettings_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.cloneSettings.$.updateOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update app clone settings", error });
                            throw error;
                        },
                    });
                },
                execute: async (request: AppCloneSettings_Execute_Req["data"]) => {
                    const result = await api.projects.apps.cloneSettings.$.execute({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to execute app clone", error });
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

export const useAppCloneSettingsApi = createHook();
