import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useTraefikServiceSettingsApi } from "~/system-settings/api/hooks";
import type {
    TraefikServiceSettings_FindOne_Req,
    TraefikServiceSettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = TraefikServiceSettings_FindOne_Req["data"];
type FindOneRes = TraefikServiceSettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useTraefikServiceSettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.traefik.service-settings.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const TraefikServiceSettingsQueries = Object.freeze({
    useFindOne,
});
