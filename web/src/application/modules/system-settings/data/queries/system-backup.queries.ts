import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useSystemBackupApi } from "~/system-settings/api/hooks";
import type { SystemBackup_FindOne_Req, SystemBackup_FindOne_Res } from "~/system-settings/api/services";
import { QK } from "~/system-settings/data/constants";

type FindOneReq = SystemBackup_FindOne_Req["data"];
type FindOneRes = SystemBackup_FindOne_Res;

function useFindOne(
    request: FindOneReq = {},
    options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useSystemBackupApi();

    return useQuery({
        queryKey: [QK["system-settings.backup.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        ...options,
    });
}

export const SystemBackupQueries = Object.freeze({
    useFindOne,
});
