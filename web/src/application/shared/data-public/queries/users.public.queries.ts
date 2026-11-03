import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useUsersPublicApi } from "@application/shared/api-public";
import type { Public_Users_FindMany_Req, Public_Users_FindMany_Res } from "@application/shared/api-public/services";
import { QK } from "@application/shared/data-public/constants";

/**
 * Find many users
 */
type FindManyReq = Public_Users_FindMany_Req["data"];
type FindManyRes = Public_Users_FindMany_Res;

type FindManyOptions = Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn">;

function useFindMany(request: FindManyReq, options: FindManyOptions = {}) {
    const { queries } = useUsersPublicApi();

    return useQuery({
        queryKey: [QK["users.public.find-many"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const UsersPublicQueries = Object.freeze({
    useFindMany,
});
