import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectsApi } from "~/projects/api";
import { type Projects_FindManyPaginated_Req, type Projects_FindManyPaginated_Res } from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

/**
 * Find many projects paginated query
 */
type FindManyPaginatedReq = Projects_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Projects_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useProjectsApi();

    return useQuery({
        queryKey: [QK["projects.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const ProjectsQueries = Object.freeze({
    useFindManyPaginated,
});
