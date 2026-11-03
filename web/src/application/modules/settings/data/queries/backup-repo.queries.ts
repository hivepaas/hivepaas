import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useBackupRepoApi } from "~/settings/api/hooks";
import type {
    BackupRepo_FindManyPaginated_Req,
    BackupRepo_FindManyPaginated_Res,
    BackupRepo_FindOneById_Req,
    BackupRepo_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = BackupRepo_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = BackupRepo_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useBackupRepoApi();

    return useQuery({
        queryKey: [QK["settings.backup-repos.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = BackupRepo_FindOneById_Req["data"];
type FindOneByIdRes = BackupRepo_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useBackupRepoApi();

    return useQuery({
        queryKey: [QK["settings.backup-repos.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

export const BackupRepoQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
