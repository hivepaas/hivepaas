import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSLoggingSettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSLoggingSettings_FindOne_Req,
    HivePaaSLoggingSettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = HivePaaSLoggingSettings_FindOne_Req["data"];
type FindOneRes = HivePaaSLoggingSettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSLoggingSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.logging.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const HivePaaSLoggingSettingsQueries = Object.freeze({
    useFindOne,
});
