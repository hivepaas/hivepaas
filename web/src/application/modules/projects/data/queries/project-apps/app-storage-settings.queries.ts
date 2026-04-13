import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppStorageSettingsApi } from "../../../api/hooks/project-apps";
import { type AppStorageSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppStorageSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppStorageSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.storage-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppStorageSettingsQueries = Object.freeze({
    useFindOne,
});
