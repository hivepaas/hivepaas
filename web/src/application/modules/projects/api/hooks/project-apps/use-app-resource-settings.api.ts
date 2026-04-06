import { use, useMemo } from "react";

import { match } from "oxide.ts";

import { useApiErrorNotifications } from "@infrastructure/api";

import { type AppResourceSettings_UpdateOne_Req } from "../../../api/services";
import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useAppResourceSettingsApi() {
        const context = use(ProjectsApiContext);
        const { api } = context;
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (request: { projectID: string; appID: string }, signal?: AbortSignal) => {
                    const result = await api.projects.apps.resourceSettings.$.findOne({ data: request }, signal);
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
                updateOne: async (request: AppResourceSettings_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.resourceSettings.$.updateOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({ message: "Failed to update resource settings", error });
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

export const useAppResourceSettingsApi = createHook();
