import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useProjectAppEnvVarsApi } from "~/projects/api/hooks/project-apps";
import type { ProjectAppEnvVars_FindOne_Req, ProjectAppEnvVars_FindOne_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find one project app env vars query
 */
type FindOneReq = ProjectAppEnvVars_FindOne_Req["data"];
type FindOneRes = ProjectAppEnvVars_FindOne_Res;

type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

function useFindOne(request: FindOneReq, options: FindOneOptions = {}) {
    const { queries } = useProjectAppEnvVarsApi();

    return useQuery({
        queryKey: [QK["projects.apps.env-vars.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const ProjectAppEnvVarsQueries = Object.freeze({
    useFindOne,
});
