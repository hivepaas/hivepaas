import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppDeploymentSettingsApi } from "../../../api/hooks/project-apps";
import { type AppDeploymentSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppDeploymentSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppDeploymentSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.deployment-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppDeploymentSettingsQueries = Object.freeze({
    useFindOne,
});
