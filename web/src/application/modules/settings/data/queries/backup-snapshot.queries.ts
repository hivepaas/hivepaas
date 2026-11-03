import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";
import { useBackupSnapshotApi } from "~/settings/api/hooks";
import type {
    BackupSnapshot_FindEntries_Req,
    BackupSnapshot_FindEntries_Res,
    BackupSnapshot_FindManyPaginated_Req,
    BackupSnapshot_FindManyPaginated_Res,
    BackupSnapshot_FindOneById_Req,
    BackupSnapshot_FindOneById_Res,
} from "~/settings/api/services";
import { QK } from "~/settings/data/constants";

type FindManyPaginatedReq = BackupSnapshot_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = BackupSnapshot_FindManyPaginated_Res;

function useFindManyPaginated(
    request: FindManyPaginatedReq,
    options: Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useBackupSnapshotApi();

    return useQuery({
        queryKey: [QK["settings.backup-snapshots.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

type FindOneByIdReq = BackupSnapshot_FindOneById_Req["data"];
type FindOneByIdRes = BackupSnapshot_FindOneById_Res;

function useFindOneById(
    request: FindOneByIdReq,
    options: Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useBackupSnapshotApi();

    return useQuery({
        queryKey: [QK["settings.backup-snapshots.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        ...options,
    });
}

type FindEntriesReq = BackupSnapshot_FindEntries_Req["data"];
type FindEntriesRes = BackupSnapshot_FindEntries_Res;

/** A directory of a snapshot: what the repository says it holds does not change. */
function useFindEntries(
    request: FindEntriesReq,
    options: Omit<UseQueryOptions<FindEntriesRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useBackupSnapshotApi();

    return useQuery({
        queryKey: [QK["settings.backup-snapshots.find-entries"], request],
        queryFn: ({ signal }) => queries.findEntries(request, signal),
        staleTime: Infinity,
        ...options,
    });
}

export const BackupSnapshotQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
    useFindEntries,
});
