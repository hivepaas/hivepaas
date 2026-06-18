import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { AppServiceTasks_FindMany_Req } from "~/projects/api/services";

function createHook() {
    return function useAppServiceTasksApi() {
        const { api } = use(ProjectsApiContext);

        const queries = useMemo(
            () => ({
                findMany: async (data: AppServiceTasks_FindMany_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.apps.serviceTasks.$.findMany({ data }, signal);

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

        return {
            queries,
        };
    };
}

export const useAppServiceTasksApi = createHook();
