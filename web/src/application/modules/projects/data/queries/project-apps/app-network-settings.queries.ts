import { type UseQueryOptions, keepPreviousData, useQuery } from "@tanstack/react-query";

import { useAppNetworkSettingsApi } from "../../../api/hooks/project-apps";
import { type AppNetworkSettings_FindOne_Res } from "../../../api/services";
import { QK } from "../../constants/projects.query-keys";

function useFindOne(
    request: { projectID: string; appID: string },
    options: Omit<UseQueryOptions<AppNetworkSettings_FindOne_Res>, "queryKey" | "queryFn"> = {},
) {
    const { queries } = useAppNetworkSettingsApi();

    return useQuery({
        queryKey: [QK["projects.apps.network-settings.$.find-one"], request],
        queryFn: ({ signal }) => queries.findOne(request, signal),
        placeholderData: keepPreviousData,
        ...options,
    });
}

export const AppNetworkSettingsQueries = Object.freeze({
    useFindOne,
});
