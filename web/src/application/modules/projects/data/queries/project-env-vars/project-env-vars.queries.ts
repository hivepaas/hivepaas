import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useProjectEnvVarsApi } from "~/projects/api";
import type { ProjectEnvVars_FindOne_Req, ProjectEnvVars_FindOne_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find one project env vars query
 */
type FindOneReq = ProjectEnvVars_FindOne_Req["data"];
type FindOneRes = ProjectEnvVars_FindOne_Res;

type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

function useFindOne(request: FindOneReq, options: FindOneOptions = {}) {
    const { queries } = useProjectEnvVarsApi();

    return useQuery({
        queryKey: [QK["projects.env-vars.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const ProjectEnvVarsQueries = Object.freeze({
    useFindOne,
});
