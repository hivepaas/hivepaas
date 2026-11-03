import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSUpdatesApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSUpdates_FindPlan_Req,
    HivePaaSUpdates_FindPlan_Res,
    HivePaaSUpdates_FindReleaseInfo_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type ReleaseInfoOptions = Omit<UseQueryOptions<HivePaaSUpdates_FindReleaseInfo_Res>, "queryKey" | "queryFn">;

/**
 * The server caches the signed release file itself, so asking often costs little.
 *
 * quiet leaves a failure unannounced, for a page that only mentions updates in
 * passing: the release file being out of reach is not news there.
 */
function useFindReleaseInfo(options: ReleaseInfoOptions = {}, { quiet = false }: { quiet?: boolean } = {}) {
    const { queries } = useHivePaaSUpdatesApi();

    return useQuery({
        // Kept apart by quiet, which only decides who hears of a failure: the
        // server caches the release file, so asking twice costs nothing.
        queryKey: [QK["system-settings.hivepaas.updates.release-info"], { quiet }],
        queryFn: ({ signal }) => queries.findReleaseInfo(signal, quiet),
        staleTime: 5 * 60_000,
        ...options,
    });
}

type PlanReq = HivePaaSUpdates_FindPlan_Req["data"];
type PlanOptions = Omit<UseQueryOptions<HivePaaSUpdates_FindPlan_Res>, "queryKey" | "queryFn">;

/** Read fresh each time it is asked: it reports the services as they run now. */
function useFindPlan(request: PlanReq, options: PlanOptions = {}) {
    const { queries } = useHivePaaSUpdatesApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.updates.plan"], request],
        queryFn: ({ signal }) => queries.findPlan(request, signal),
        enabled: Boolean(request.targetVersion),
        staleTime: 0,
        ...options,
    });
}

export const HivePaaSUpdatesQueries = Object.freeze({
    useFindReleaseInfo,
    useFindPlan,
});
