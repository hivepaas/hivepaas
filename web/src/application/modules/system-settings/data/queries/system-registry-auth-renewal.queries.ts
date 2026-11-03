import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSystemRegistryAuthRenewalApi } from "~/system-settings/api/hooks";
import type {
    SystemRegistryAuthRenewal_FindOne_Req,
    SystemRegistryAuthRenewal_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = SystemRegistryAuthRenewal_FindOne_Req["data"];
type FindOneRes = SystemRegistryAuthRenewal_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useSystemRegistryAuthRenewalApi();

    return useQuery({
        queryKey: [QK["system-settings.registry-auth-renewal.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const SystemRegistryAuthRenewalQueries = Object.freeze({
    useFindOne,
});
