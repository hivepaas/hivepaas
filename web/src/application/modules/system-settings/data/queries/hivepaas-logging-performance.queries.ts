import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSLoggingPerformanceApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSLoggingPerformance_FindOne_Req,
    HivePaaSLoggingPerformance_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = HivePaaSLoggingPerformance_FindOne_Req["data"];
type FindOneRes = HivePaaSLoggingPerformance_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSLoggingPerformanceApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.logging-performance.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const HivePaaSLoggingPerformanceQueries = Object.freeze({
    useFindOne,
});
