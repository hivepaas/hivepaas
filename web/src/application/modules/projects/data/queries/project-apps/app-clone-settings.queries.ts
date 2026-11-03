import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppCloneSettingsApi } from "../../../api/hooks/project-apps";
import type { AppCloneSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; env: string; appID: string },
    options: Omit<UseQueryOptions<AppCloneSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppCloneSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.clone-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppCloneSettingsQueries = Object.freeze({
    useFindOne,
});
