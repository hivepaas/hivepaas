import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useAccessTokenApi } from "~/settings/api/hooks";
import type {
    AccessToken_FindManyPaginated_Req,
    AccessToken_FindManyPaginated_Res,
    AccessToken_FindOneById_Req,
    AccessToken_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = AccessToken_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = AccessToken_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAccessTokenApi();

    return useQuery({
        queryKey: [QK["settings.access-token.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = AccessToken_FindOneById_Req["data"];
type FindOneByIdRes = AccessToken_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAccessTokenApi();

    return useQuery({
        queryKey: [QK["settings.access-token.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const AccessTokenQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
