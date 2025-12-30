import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { Projects_FindManyPaginated_Req } from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectsApi() {
        const { api } = use(ProjectsApiContext);

        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                /**
                 * Find many projects paginated
                 */
                findManyPaginated: async (data: Projects_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get projects",
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
        };
    };
}

export const useProjectsApi = createHook();
