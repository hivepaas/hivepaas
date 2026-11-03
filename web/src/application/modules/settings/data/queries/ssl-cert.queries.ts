import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useSslCertApi } from "~/settings/api/hooks";
import type {
    SslCert_FindManyPaginated_Req,
    SslCert_FindManyPaginated_Res,
    SslCert_FindOneById_Req,
    SslCert_FindOneById_Res,
} from "~/settings/api/services/ssl-cert-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = SslCert_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = SslCert_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSslCertApi();

    return useQuery({
        queryKey: [QK["settings.ssl-cert.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = SslCert_FindOneById_Req["data"];
type FindOneByIdRes = SslCert_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSslCertApi();

    return useQuery({
        queryKey: [QK["settings.ssl-cert.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const SslCertQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
