import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppContainerSettingsApi } from "../../../api/hooks/project-apps";
import { type AppContainerSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppContainerSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppContainerSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.container-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppContainerSettingsQueries = Object.freeze({
    useFindOne,
});
