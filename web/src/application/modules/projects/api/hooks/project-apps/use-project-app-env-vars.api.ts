import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { ProjectAppEnvVars_FindOne_Req, ProjectAppEnvVars_UpdateOne_Req } from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectAppEnvVarsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find one project app env vars
                 */
                findOne: async (data: ProjectAppEnvVars_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.envVars.$.findOne(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get app env vars",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        const mutations = useMemo(
            () => ({
                /**
                 * Update project app env vars
                 */
                updateOne: async (data: ProjectAppEnvVars_UpdateOne_Req["data"]) => {
                    const result = await api.projects.apps.envVars.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update app env vars",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return {
            queries,
            mutations,
        };
    };
}

export const useProjectAppEnvVarsApi = createHook();
