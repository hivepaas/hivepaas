import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useTraefikConfigOptionsApi } from "~/system-settings/api/hooks";
import type {
    TraefikConfigOptions_FindOne_Req,
    TraefikConfigOptions_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = TraefikConfigOptions_FindOne_Req["data"];
type FindOneRes = TraefikConfigOptions_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useTraefikConfigOptionsApi();

    return useQuery({
        queryKey: [QK["system-settings.traefik.config-options.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

/**
 * Polls the config options endpoint while a change is on trial.
 *
 * Two answers come out of one request: whether the caller can still reach
 * HivePaaS at all - which here means whether the new Traefik is routing anything
 * - and whether the trial is still running.
 */
function useProbe(options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useTraefikConfigOptionsApi();

    return useQuery({
        queryKey: [QK["system-settings.traefik.config-options.probe"]],
        queryFn: ({ signal }) => queries.probe(signal),
        // A failed probe is information, not something to paper over: retrying
        // inside one tick would blur the boundary between "slow" and "locked
        // out", which is the only thing this query exists to tell apart.
        retry: false,
        gcTime: 0,
        ...options,
    });
}

export const TraefikConfigOptionsQueries = Object.freeze({
    useFindOne,
    useProbe,
});
