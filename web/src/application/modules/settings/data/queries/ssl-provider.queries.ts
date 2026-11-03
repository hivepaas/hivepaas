import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useSslProviderApi } from "~/settings/api/hooks";
import type {
    SslProvider_FindManyPaginated_Req,
    SslProvider_FindManyPaginated_Res,
    SslProvider_FindOneById_Req,
    SslProvider_FindOneById_Res,
} from "~/settings/api/services/ssl-provider-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = SslProvider_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = SslProvider_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSslProviderApi();

    return useQuery({
        queryKey: [QK["settings.ssl-provider.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = SslProvider_FindOneById_Req["data"];
type FindOneByIdRes = SslProvider_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSslProviderApi();

    return useQuery({
        queryKey: [QK["settings.ssl-provider.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const SslProviderQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
