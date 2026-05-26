import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useProjectUserAccessesApi } from "~/projects/api";
import type { ProjectUserAccesses_FindOne_Req, ProjectUserAccesses_FindOne_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find project user accesses query
 */
type FindOneReq = ProjectUserAccesses_FindOne_Req["data"];
type FindOneRes = ProjectUserAccesses_FindOne_Res;

type FindOneOptions = Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn">;

function useFindOne(request: FindOneReq, options: FindOneOptions = {}) {
    const { queries } = useProjectUserAccessesApi();

    return useQuery({
        queryKey: [QK["projects.user-accesses.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const ProjectUserAccessesQueries = Object.freeze({
    useFindOne,
});
