import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectSslProviderApi } from "~/projects/api/hooks";
import type {
    ProjectSslProvider_FindManyPaginated_Req,
    ProjectSslProvider_FindManyPaginated_Res,
    ProjectSslProvider_FindOneById_Req,
    ProjectSslProvider_FindOneById_Res,
} from "~/projects/api/services";
import { PROJECTS_LIST_QUERY_OPTIONS, QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectSslProvider_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectSslProvider_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectSslProviderApi();

    return useQuery({
        queryKey: [QK["projects.ssl-provider.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...PROJECTS_LIST_QUERY_OPTIONS,
        ...options,
    });
}

type FindOneByIdReq = ProjectSslProvider_FindOneById_Req["data"];
type FindOneByIdRes = ProjectSslProvider_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectSslProviderApi();

    return useQuery({
        queryKey: [QK["projects.ssl-provider.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectSslProviderQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
