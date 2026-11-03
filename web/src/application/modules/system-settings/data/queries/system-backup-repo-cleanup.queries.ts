import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSystemBackupRepoCleanupApi } from "~/system-settings/api/hooks";
import type {
    SystemBackupRepoCleanup_FindOne_Req,
    SystemBackupRepoCleanup_FindOne_Res,
} from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = SystemBackupRepoCleanup_FindOne_Req["data"];
type FindOneRes = SystemBackupRepoCleanup_FindOne_Res;

function useFindOne(request: FindOneReq = {}, options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useSystemBackupRepoCleanupApi();

    return useQuery({
        queryKey: [QK["system-settings.backup-repo-cleanup.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const SystemBackupRepoCleanupQueries = Object.freeze({
    useFindOne,
});
