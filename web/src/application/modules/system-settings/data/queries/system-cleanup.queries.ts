import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSystemCleanupApi } from "~/system-settings/api/hooks";
import type { SystemCleanup_FindOne_Req, SystemCleanup_FindOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = SystemCleanup_FindOne_Req["data"];
type FindOneRes = SystemCleanup_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useSystemCleanupApi();

    return useQuery({
        queryKey: [QK["system-settings.cleanup.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const SystemCleanupQueries = Object.freeze({
    useFindOne,
});
