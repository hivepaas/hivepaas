import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useHivePaaSSecuritySettingsApi } from "~/system-settings/api/hooks";
import type {
    HivePaaSSecuritySettings_FindOne_Req,
    HivePaaSSecuritySettings_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = HivePaaSSecuritySettings_FindOne_Req["data"];
type FindOneRes = HivePaaSSecuritySettings_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useHivePaaSSecuritySettingsApi();

    return useQuery({
        queryKey: [QK["system-settings.hivepaas.security-settings.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const HivePaaSSecuritySettingsQueries = Object.freeze({
    useFindOne,
});
