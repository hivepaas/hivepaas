import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSystemBackupFileApi } from "~/system-settings/api/hooks";
import type {
    SystemBackupFile_FindManyPaginated_Req,
    SystemBackupFile_FindManyPaginated_Res,
    SystemBackupFile_FindOneById_Req,
    SystemBackupFile_FindOneById_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindManyPaginatedReq = SystemBackupFile_FindManyPaginated_Req["data"];
type FindManyPaginatedRes = SystemBackupFile_FindManyPaginated_Res;
type FindManyPaginatedOptions = Omit<UseQueryOptions<FindManyPaginatedRes>, "queryKey" | "queryFn">;

type FindOneByIdReq = SystemBackupFile_FindOneById_Req["data"];
type FindOneByIdRes = SystemBackupFile_FindOneById_Res;
type FindOneByIdOptions = Omit<UseQueryOptions<FindOneByIdRes>, "queryKey" | "queryFn">;

function useFindManyPaginated(request: FindManyPaginatedReq = {}, options: FindManyPaginatedOptions = {}) {
    const { queries } = useSystemBackupFileApi();

    return useQuery({
        queryKey: [QK["system-settings.backup-files.find-many-paginated"], request],
        queryFn: ({ signal }) => queries.findManyPaginated(request, signal),
        ...options,
    });
}

function useFindOneById(request: FindOneByIdReq, options: FindOneByIdOptions = {}) {
    const { queries } = useSystemBackupFileApi();

    return useQuery({
        queryKey: [QK["system-settings.backup-files.find-one-by-id"], request],
        queryFn: ({ signal }) => queries.findOneById(request, signal),
        enabled: Boolean(request.fileID),
        ...options,
    });
}

export const SystemBackupFileQueries = Object.freeze({
    useFindManyPaginated,
    useFindOneById,
});
