import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppRoutingSettingsApi } from "../../../api/hooks/project-apps";
import { type AppRoutingSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; env: string; appID: string },
    options: Omit<UseQueryOptions<AppRoutingSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppRoutingSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.routing-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppRoutingSettingsQueries = Object.freeze({
    useFindOne,
});

export { AppRoutingSettingsQueries as AppHttpSettingsQueries };
