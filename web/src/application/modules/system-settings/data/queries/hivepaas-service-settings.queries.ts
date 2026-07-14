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

export const HivePaaSServiceSettingsQueries = Object.freeze({
    useFindOne,
});
