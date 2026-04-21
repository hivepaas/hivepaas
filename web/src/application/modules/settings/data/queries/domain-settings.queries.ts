import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useDomainSettingsApi } from "~/settings/api/hooks";
import type {
    DomainSettings_FindOne_Req,
    DomainSettings_FindOne_Res,
} from "~/settings/api/services/domain-settings-services";
import { QK } from "~/settings/data/constants";

type FindOneReq = DomainSettings_FindOne_Req["data"];
type FindOneRes = DomainSettings_FindOne_Res;

function useFindOne(
    request: FindOneReq = {},
    options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useDomainSettingsApi();

    return useQuery({
        queryKey: [QK["settings.domain-settings.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const DomainSettingsQueries = Object.freeze({
    useFindOne,
});
