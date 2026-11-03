import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppDockerApiSettingsApi } from "../../../api/hooks/project-apps";
import type { AppDockerApiSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; env: string; appID: string },
    options: Omit<UseQueryOptions<AppDockerApiSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppDockerApiSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.docker-api-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppDockerApiSettingsQueries = Object.freeze({
    useFindOne,
});
