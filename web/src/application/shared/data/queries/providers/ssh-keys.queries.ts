import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useSshKeysApi } from "@application/shared/api/hooks/providers";
import {
    type SshKeys_FindManyPaginated_Req,
    type SshKeys_FindManyPaginated_Res,
} from "@application/shared/api/services";
import { QK } from "@application/shared/data/constants";

/**
 * Find many SSH keys paginated query
 */
type FindManyPaginatedReq = SshKeys_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = SshKeys_FindManyPaginated_Res;

type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useSshKeysApi();

    return useQuery({
        queryKey: [QK["ssh-keys.$.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const SshKeysPublicQueries = Object.freeze({
    useFindManyPaginated,
});
