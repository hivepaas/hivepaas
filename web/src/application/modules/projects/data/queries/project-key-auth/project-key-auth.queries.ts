import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectKeyAuthApi } from "~/projects/api/hooks";
import type {
    ProjectKeyAuth_FindManyPaginated_Req,
    ProjectKeyAuth_FindManyPaginated_Res,
    ProjectKeyAuth_FindOneById_Req,
    ProjectKeyAuth_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectKeyAuth_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectKeyAuth_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectKeyAuthApi();

    return useQuery({
        queryKey: [QK["projects.key-auth.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = ProjectKeyAuth_FindOneById_Req["data"];
type FindOneByIdRes = ProjectKeyAuth_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectKeyAuthApi();

    return useQuery({
        queryKey: [QK["projects.key-auth.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectKeyAuthQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
