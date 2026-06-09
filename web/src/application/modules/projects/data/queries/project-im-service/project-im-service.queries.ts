import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectImServiceApi } from "~/projects/api/hooks";
import type {
    ProjectImService_FindManyPaginated_Req,
    ProjectImService_FindManyPaginated_Res,
    ProjectImService_FindOneById_Req,
    ProjectImService_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectImService_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectImService_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectImServiceApi();

    return useQuery({
        queryKey: [QK["projects.im-service.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = ProjectImService_FindOneById_Req["data"];
type FindOneByIdRes = ProjectImService_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectImServiceApi();

    return useQuery({
        queryKey: [QK["projects.im-service.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectImServiceQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
