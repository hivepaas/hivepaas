import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppResourceSettingsApi } from "../../../api/hooks/project-apps";
import { type AppResourceSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppResourceSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppResourceSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.resource-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppResourceSettingsQueries = Object.freeze({
    useFindOne,
});
