import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useLocalPaaSServiceSettingsApi } from "~/system-settings/api/hooks";
import type {
    LocalPaaSServiceSettings_FindOne_Req,
    LocalPaaSServiceSettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = LocalPaaSServiceSettings_FindOne_Req["data"];
type FindOneRes = LocalPaaSServiceSettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useLocalPaaSServiceSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.localpaas.service-settings.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const LocalPaaSServiceSettingsQueries = Object.freeze({
    useFindOne,
});
