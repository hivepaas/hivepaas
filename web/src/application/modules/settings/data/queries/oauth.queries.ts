import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useOAuthApi } from "~/settings/api/hooks";
import type {
    OAuth_FindManyPaginated_Req,
    OAuth_FindManyPaginated_Res,
    OAuth_FindOneById_Req,
    OAuth_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = OAuth_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = OAuth_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useOAuthApi();

    return useQuery({
        queryKey: [QK["settings.oauth.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = OAuth_FindOneById_Req["data"];
type FindOneByIdRes = OAuth_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useOAuthApi();

    return useQuery({
        queryKey: [QK["settings.oauth.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const OAuthQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
