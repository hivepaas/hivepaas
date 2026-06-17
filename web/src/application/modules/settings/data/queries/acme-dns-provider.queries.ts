import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAcmeDnsProviderApi } from "~/settings/api/hooks";
import type {
    AcmeDnsProvider_FindManyPaginated_Req,
    AcmeDnsProvider_FindManyPaginated_Res,
    AcmeDnsProvider_FindOneById_Req,
    AcmeDnsProvider_FindOneById_Res,
} from "~/settings/api/services/acme-dns-provider-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = AcmeDnsProvider_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AcmeDnsProvider_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAcmeDnsProviderApi();

    return useQuery({
        queryKey: [QK["settings.acme-dns-provider.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = AcmeDnsProvider_FindOneById_Req["data"];
type FindOneByIdRes = AcmeDnsProvider_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAcmeDnsProviderApi();

    return useQuery({
        queryKey: [QK["settings.acme-dns-provider.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const AcmeDnsProviderQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
