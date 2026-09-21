import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSRegistrySettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSRegistrySettings_FindOne_Req,
    HivePaaSRegistrySettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = HivePaaSRegistrySettings_FindOne_Req["data"];
type FindOneRes = HivePaaSRegistrySettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSRegistrySettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.registry.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const HivePaaSRegistrySettingsQueries = Object.freeze({
    useFindOne,
});
