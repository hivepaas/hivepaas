import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppsPublicApi } from "@application/shared/api-public";
import type { Public_Apps_FindMany_Req, Public_Apps_FindMany_Res } from "@application/shared/api-public/services";
import { QK } from "@application/shared/data-public/constants";

/**
 * Find many apps
 */
type FindManyReq = Public_Apps_FindMany_Req["data"];
type FindManyRes = Public_Apps_FindMany_Res;

type FindManyOptions = Omit<UseQueryOptions<FindManyRes>, "queryKey" | "queryFn">;

function useFindMany(request: FindManyReq, options: FindManyOptions = {}) {
    const { queries } = useAppsPublicApi();

    return useQuery({
        queryKey: [QK["apps.public.find-many"], request],
        queryFn: ({ signal }) => queries.findMany(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppsPublicQueries = Object.freeze({
    useFindMany,
});
