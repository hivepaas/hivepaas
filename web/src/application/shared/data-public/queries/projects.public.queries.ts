import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useProjectsPublicApi } from "@application/shared/api-public";
import type {
    Public_Projects_FindManyPaginated_Req,
    Public_Projects_FindManyPaginated_Res,
} from "@application/shared/api-public/services";
import { QK } from "@application/shared/data-public/constants";

/**
 * Find many paginated projects
 */
type FindManyPaginatedReq = Public_Projects_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Public_Projects_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectsPublicApi();

    return useQuery({
        queryKey: [QK["projects.public.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,

        ...options,
    });
}

export const ProjectsPublicQueries = Object.freeze({
    useFindManyPaginated,
});
