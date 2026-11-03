import { type UseQueryOptions, useQuery } from "@tanstack/react-query";
import { useStorageSettingsApi } from "~/settings/api/hooks";
import type { StorageSettings_FindOne_Res } from "~/settings/api/services/storage-settings-services";
import { QK } from "~/settings/data/constants";

type FindOneRes = StorageSettings_FindOne_Res;

function useFindOne(options: Omit<UseQueryOptions<FindOneRes>, "queryKey" | "queryFn"> = {}) {
    const { queries } = useStorageSettingsApi();

    return useQuery({
        queryKey: [QK["settings.storage-settings.find-one"]],
        queryFn: ({ signal }) => queries.findOne(signal),
        ...options,
    });
}

export const StorageSettingsQueries = Object.freeze({
    useFindOne,
});
