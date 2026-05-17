import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useSSHKeyApi } from "~/settings/api/hooks";
import type {
    SSHKey_FindManyPaginated_Req,
    SSHKey_FindManyPaginated_Res,
    SSHKey_FindOneById_Req,
    SSHKey_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = SSHKey_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = SSHKey_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSSHKeyApi();

    return useQuery({
        queryKey: [QK["settings.ssh-key.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = SSHKey_FindOneById_Req["data"];
type FindOneByIdRes = SSHKey_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSSHKeyApi();

    return useQuery({
        queryKey: [QK["settings.ssh-key.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const SSHKeyQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
