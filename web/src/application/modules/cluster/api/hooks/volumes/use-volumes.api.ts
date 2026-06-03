import { use, useMemo } from "react";

import { match } from "oxide.ts";
import { NodesApiContext } from "~/cluster/api/api-context";
import type { ClusterVolumes_List_Req } from "~/cluster/api/services";

import { useApiErrorNotifications } from "@infrastructure/api";

function createHook() {
    return function useClusterVolumesApi() {
        const { api } = use(NodesApiContext);
        const { notifyError } = useApiErrorNotifications();

        const queries = useMemo(
            () => ({
                list: async (data: ClusterVolumes_List_Req["data"], signal?: AbortSignal) => {
                    const result = await api.volumes.$.list({ data }, signal);

                    return match(result, {
                        Ok: _ => _,
                        Err: error => {
                            notifyError({
                                message: "Failed to get cluster volumes",
                                error,
                            });

                            throw error;
                        },
                    });
                },
            }),
            [api, notifyError],
        );

        return { queries };
    };
}

export const useClusterVolumesApi = createHook();
