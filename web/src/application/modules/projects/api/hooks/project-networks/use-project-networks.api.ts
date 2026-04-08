import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { ProjectsApiContext } from "~/projects/api/api-context";
import type { ProjectNetworks_FindManyPaginated_Req } from "~/projects/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useProjectNetworksApi() {
        const { api } = use(ProjectsApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                findManyPaginated: async (data: ProjectNetworks_FindManyPaginated_Req["data"], signal?: AbortSignal) => {
                    const result = await api.projects.networks.$.findManyPaginated(
                        {
                            data,
                        },
                        signal,
                    );

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get project networks",
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

export const useProjectNetworksApi = createHook();
