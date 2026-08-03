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

export const TraefikConfigOptionsQueries = Object.freeze({
    useFindOne,
});
