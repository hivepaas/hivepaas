import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useUsersApi } from "~/user-management/api";
import { type Users_FindManyPaginated_Req, type Users_FindManyPaginated_Res } from "~/user-management/api/services";
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

export const UsersQueries = Object.freeze({
    useFindManyPaginated,
});
