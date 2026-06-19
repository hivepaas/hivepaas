import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppFeatureSettingsApi } from "../../../api/hooks/project-apps";
import type { AppFeatureSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppFeatureSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppFeatureSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.feature-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppFeatureSettingsQueries = Object.freeze({
    useFindOne,
});
