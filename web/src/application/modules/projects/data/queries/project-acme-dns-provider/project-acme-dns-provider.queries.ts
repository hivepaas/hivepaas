import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useProjectAcmeDnsProviderApi } from "~/projects/api/hooks";
import type {
    ProjectAcmeDnsProvider_FindManyPaginated_Req,
    ProjectAcmeDnsProvider_FindManyPaginated_Res,
    ProjectAcmeDnsProvider_FindOneById_Req,
    ProjectAcmeDnsProvider_FindOneById_Res,
} from "~/projects/api/services";
import { QK } from "~/projects/data/constants";

type FindManyPaginatedReq = ProjectAcmeDnsProvider_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = ProjectAcmeDnsProvider_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectAcmeDnsProviderApi();

    return useQuery({
        queryKey: [QK["projects.acme-dns-provider.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = ProjectAcmeDnsProvider_FindOneById_Req["data"];
type FindOneByIdRes = ProjectAcmeDnsProvider_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useProjectAcmeDnsProviderApi();

    return useQuery({
        queryKey: [QK["projects.acme-dns-provider.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const ProjectAcmeDnsProviderQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
