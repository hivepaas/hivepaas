import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useBasicAuthApi } from "~/settings/api/hooks";
import type {
    BasicAuth_FindManyPaginated_Req,
    BasicAuth_FindManyPaginated_Res,
    BasicAuth_FindOneById_Req,
    BasicAuth_FindOneById_Res,
} from "~/settings/api/services/basic-auth-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = BasicAuth_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = BasicAuth_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useBasicAuthApi();

    return useQuery({
        queryKey: [QK["settings.basic-auth.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = BasicAuth_FindOneById_Req["data"];
type FindOneByIdRes = BasicAuth_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useBasicAuthApi();

    return useQuery({
        queryKey: [QK["settings.basic-auth.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const BasicAuthQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
