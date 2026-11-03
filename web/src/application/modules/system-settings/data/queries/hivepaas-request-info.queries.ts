import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSRequestInfoApi } from "~/system-settings/api/hooks";
import type { HivePaaSRequestInfo_FindOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneRes = HivePaaSRequestInfo_FindOne_Res;

/**
 * Measures how this browser's own request reached HivePaaS.
 *
 * The answer describes the path the caller took, so it is only worth anything
 * when the caller is real traffic - which a dashboard in somebody's browser is.
 * It says nothing useful when read from inside the cluster.
 */
function useFindOne(options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSRequestInfoApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.request-info.find-one"]],
        queryFn: ({ signal }) => queries.findOne({}, signal),
        // A measurement, not state: taken once when the form opens, and not worth
        // retrying or refreshing behind the operator's back.
        retry: false,
        refetchOnWindowFocus: false,
        staleTime: Infinity,
        ...options,
    });
}

export const HivePaaSRequestInfoQueries = Object.freeze({
    useFindOne,
});
