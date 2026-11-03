import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useProfileApi } from "@application/shared/api/hooks";
import {
    type Profile_FindManyApiKeysPaginated_Req,
    type Profile_FindManyApiKeysPaginated_Res,
} from "@application/shared/api/services";
import { QK } from "@application/shared/data/constants";

/**
 * Find many account API keys paginated query
 */
type FindManyApiKeysPaginatedReq = Profile_FindManyApiKeysPaginated_Req["data"];
type FindManyApiKeysPaginatedRes = Profile_FindManyApiKeysPaginated_Res;

type FindManyApiKeysPaginatedOptions = Omit<UseQueryOptions<FindManyApiKeysPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyApiKeysPaginated(
    request: FindManyApiKeysPaginatedReq = {},
    options: FindManyApiKeysPaginatedOptions = {},
) {
    const { queries } = useProfileApi();

    return useQuery({
        queryKey: [QK["profile-api-keys.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyApiKeysPaginated(request, signal),
        placeholderData: keepPreviousData,

        ...options,
    });
}

export const ProfileQueries = Object.freeze({
    useFindManyApiKeysPaginated,
});
