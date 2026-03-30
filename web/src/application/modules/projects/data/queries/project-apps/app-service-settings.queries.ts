import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppServiceSettingsApi } from "../../../api/hooks/project-apps";
import { type AppServiceSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppServiceSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppServiceSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.service-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppServiceSettingsQueries = Object.freeze({
    useFindOne,
});
