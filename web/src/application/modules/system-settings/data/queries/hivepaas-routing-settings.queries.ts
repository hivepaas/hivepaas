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

export const HivePaaSRoutingSettingsQueries = Object.freeze({
    useFindOne,
});

export { HivePaaSRoutingSettingsQueries as HivePaaSHttpSettingsQueries };
