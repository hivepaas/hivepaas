import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSServiceSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSServiceSettings_FindOne_Req,
    HivePaaSServiceSettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = HivePaaSServiceSettings_FindOne_Req["data"];
type FindOneRes = HivePaaSServiceSettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSServiceSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.service-settings.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

/**
 * Polls the settings endpoint while a proxy change is on trial.
 *
 * Two answers from one request: whether the caller can reach HivePaaS at all,
 * and whether the trial is still running. During a proxy change the first answer
 * is "no" for a while by design - traefik is being restarted - which is why the
 * dialog reads it against confirmableFrom rather than on its own.
 */
function useProbe(options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSServiceSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.service-settings.probe"]],
        queryFn: ({ signal }) => queries.probe(signal),
        retry: false,
        gcTime: 0,
        ...options,
    });
}

export const HivePaaSServiceSettingsQueries = Object.freeze({
    useFindOne,
    useProbe,
});
