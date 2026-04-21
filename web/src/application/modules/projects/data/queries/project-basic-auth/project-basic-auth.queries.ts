import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectBasicAuthApi } from "~/projects/api/hooks";
import type {
    ProjectBasicAuth_FindManyPaginated_Req,
    ProjectBasicAuth_FindManyPaginated_Res,
    ProjectBasicAuth_FindOneById_Req,
    ProjectBasicAuth_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectBasicAuth_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectBasicAuth_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectBasicAuthApi();

    return useQuery({
        queryKey: [QK["projects.basic-auth.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectBasicAuth_FindOneById_Req["data"];
type FindOneByIdRes = ProjectBasicAuth_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectBasicAuthApi();

    return useQuery({
        queryKey: [QK["projects.basic-auth.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectBasicAuthQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
