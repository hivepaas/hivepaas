import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSHttpSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSHttpSettings_FindOne_Req,
    HivePaaSHttpSettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = HivePaaSHttpSettings_FindOne_Req["data"];
type FindOneRes = HivePaaSHttpSettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSHttpSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.http-settings.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const HivePaaSHttpSettingsQueries = Object.freeze({
    useFindOne,
});
