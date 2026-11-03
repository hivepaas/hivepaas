import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useUsersApi } from "~/user-management/api";
import {
    type Users_FindManyPaginated_Req,
    type Users_FindManyPaginated_Res,
    type Users_FindOneById_Req,
    type Users_FindOneById_Res,
} from "~/user-management/api/services";
import { QK } from "~/user-management/data/constants";

/**
 * Find many users paginated query
 */
type FindManyPaginatedReq = Users_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = Users_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useUsersApi();

    return useQuery({
        queryKey: [QK["users.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,

        ...options,
    });
}

/**
 * Find one user by id query
 */
type FindOneByIdReq = Users_FindOneById_Req["data"];
type FindOneByIdRes = Users_FindOneById_Res;

type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useUsersApi();

    return useQuery({
        queryKey: [QK["users.$.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const UsersQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
