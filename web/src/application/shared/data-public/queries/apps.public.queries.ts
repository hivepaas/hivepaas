import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppsPublicApi } from "@application/shared/api-public";
import type {
    Public_Apps_FindManyBase_Req,
    Public_Apps_FindManyBase_Res,
} from "@application/shared/api-public/services";
import { QK } from "@application/shared/data-public/constants";

/**
 * Find many base apps
 */
type FindManyBaseReq = Public_Apps_FindManyBase_Req["data"];
type FindManyBaseRes = Public_Apps_FindManyBase_Res;

type FindManyBaseOptions = Omit<UseQueryOptions<FindManyBaseRes>, "queryKey" | "queryFn">;

function useFindManyBase(request: FindManyBaseReq, options: FindManyBaseOptions = {}) {
    const { queries } = useAppsPublicApi();

    return useQuery({
        queryKey: [QK["apps.public.find-many-base"], request],
        queryFn: ({ signal }) => queries.findManyBase(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppsPublicQueries = Object.freeze({
    useFindManyBase,
});
