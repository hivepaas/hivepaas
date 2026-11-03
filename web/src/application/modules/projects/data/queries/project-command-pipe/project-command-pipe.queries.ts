import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectCommandPipeApi } from "~/projects/api/hooks";
import type {
    ProjectCommandPipe_FindManyPaginated_Req,
    ProjectCommandPipe_FindManyPaginated_Res,
    ProjectCommandPipe_FindOneById_Req,
    ProjectCommandPipe_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectCommandPipe_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectCommandPipe_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectCommandPipeApi();

    return useQuery({
        queryKey: [QK["projects.command-pipes.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = ProjectCommandPipe_FindOneById_Req["data"];
type FindOneByIdRes = ProjectCommandPipe_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectCommandPipeApi();

    return useQuery({
        queryKey: [QK["projects.command-pipes.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectCommandPipeQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
