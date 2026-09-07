import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSRoutingSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSRoutingSettings_FindOne_Req,
    HivePaaSRoutingSettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = HivePaaSRoutingSettings_FindOne_Req["data"];
type FindOneRes = HivePaaSRoutingSettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSRoutingSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.routing-settings.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

/**
 * Polls the settings endpoint while a change is on trial.
 *
 * Two answers come out of one request: whether the caller can still reach
 * HivePaaS at all, and whether the trial is still running. When it stops
 * returning a pendingChange, the change has been confirmed, reverted, or undone
 * at its deadline.
 */
function useProbe(options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSRoutingSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.routing-settings.probe"]],
        queryFn: ({ signal }) => queries.probe(signal),
        // A failed probe is information, not something to paper over: retrying
        // inside one tick would blur the boundary between "slow" and "locked
        // out", which is the only thing this query exists to tell apart.
        retry: false,
        gcTime: 0,
        ...options,
    });
}

export const HivePaaSRoutingSettingsQueries = Object.freeze({
    useFindOne,
    useProbe,
});

export { HivePaaSRoutingSettingsQueries as HivePaaSHttpSettingsQueries };
