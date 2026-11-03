import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSystemSslRenewalApi } from "~/system-settings/api/hooks";
import type { SystemSslRenewal_FindOne_Req, SystemSslRenewal_FindOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = SystemSslRenewal_FindOne_Req["data"];
type FindOneRes = SystemSslRenewal_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useSystemSslRenewalApi();

    return useQuery({
        queryKey: [QK["system-settings.ssl-renewal.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const SystemSslRenewalQueries = Object.freeze({
    useFindOne,
});
