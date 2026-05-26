import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { ProjectUserAccesses_FindOne_Req, ProjectUserAccesses_UpdateOne_Req } from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectUserAccessesApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find project user accesses
                 */
                findOne: async (data: ProjectUserAccesses_FindOne_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.userAccesses.$.findOne(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project user accesses",
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
                 * Update project user accesses
                 */
                updateOne: async (data: ProjectUserAccesses_UpdateOne_Req["data"]) => {
                    const result = await api.projects.userAccesses.$.updateOne({
                        data,
                    });

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to update project user accesses",
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

export const useProjectUserAccessesApi = createHook();
