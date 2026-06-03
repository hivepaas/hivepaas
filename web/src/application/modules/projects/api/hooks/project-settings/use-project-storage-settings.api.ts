import { use, useMemo } from "react";

import { match } from "oxide.ts";
import type { ProjectStorageSettings_UpdateOne_Req } from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

import { ProjectsApiContext } from "../../api-context/projects.api.context";

function createHook() {
    return function useProjectStorageSettingsApi() {
        const context = use(ProjectsApiContext);
        const { api } = context;
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findOne: async (request: { projectID: string }, signal?: AbortSignal) => {
                    const result = await api.projects.storageSettings.$.findOne({ data: request }, signal);
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
                updateOne: async (request: ProjectStorageSettings_UpdateOne_Req["data"]) => {
                    const result = await api.projects.storageSettings.$.updateOne({ data: request });
                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project storage settings",
                                error,
                            });

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

export const useProjectStorageSettingsApi = createHook();
