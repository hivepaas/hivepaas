import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useProjectDomainSettingsApi } from "~/projects/api/hooks";
import type {
    ProjectDomainSettings_FindOne_Req,
    ProjectDomainSettings_FindOne_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindOneReq = ProjectDomainSettings_FindOne_Req["data"];
type FindOneRes = ProjectDomainSettings_FindOne_Res;

function useFindOne(
    request: FindOneReq,
    options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectDomainSettingsApi();

    return useQuery({
        queryKey: [QK["projects.domain-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const ProjectDomainSettingsQueries = Object.freeze({
    useFindOne,
});
