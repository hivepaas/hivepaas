import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useKeyAuthApi } from "~/settings/api/hooks";
import type {
    KeyAuth_FindManyPaginated_Req,
    KeyAuth_FindManyPaginated_Res,
    KeyAuth_FindOneById_Req,
    KeyAuth_FindOneById_Res,
} from "~/settings/api/services/key-auth-services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = KeyAuth_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = KeyAuth_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useKeyAuthApi();

    return useQuery({
        queryKey: [QK["settings.key-auth.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = KeyAuth_FindOneById_Req["data"];
type FindOneByIdRes = KeyAuth_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useKeyAuthApi();

    return useQuery({
        queryKey: [QK["settings.key-auth.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const KeyAuthQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
